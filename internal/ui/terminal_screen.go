package ui

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	defaultTerminalWidth  = 80
	defaultTerminalHeight = 24
	maxTerminalWidth      = 240
	maxTerminalHeight     = 100
)

type terminalScreen struct {
	cells         [][]string
	row, col      int
	savedRow      int
	savedCol      int
	width, height int
	scrollTop     int
	scrollBottom  int
	wrapPending   bool
}

func terminalScreenSnapshot(data []byte) string {
	screen := newTerminalScreen(defaultTerminalWidth, defaultTerminalHeight)
	for index := 0; index < len(data); {
		if data[index] == 0x1b {
			index = screen.consumeEscape(data, index)
			continue
		}
		switch data[index] {
		case '\r':
			screen.col = 0
			screen.wrapPending = false
			index++
			continue
		case '\n':
			screen.wrapPending = false
			screen.lineFeed()
			index++
			continue
		case '\b':
			screen.wrapPending = false
			if screen.col > 0 {
				screen.col--
			}
			index++
			continue
		case '\t':
			screen.wrapPending = false
			screen.col = minInt(((screen.col/8)+1)*8, screen.width-1)
			index++
			continue
		}
		r, size := utf8.DecodeRune(data[index:])
		if r == utf8.RuneError && size == 1 {
			index++
			continue
		}
		if !unicode.IsControl(r) {
			screen.writeRune(r)
		}
		index += size
	}
	return screen.String()
}

func newTerminalScreen(width, height int) *terminalScreen {
	screen := &terminalScreen{width: width, height: height, scrollBottom: height - 1}
	screen.cells = makeCells(width, height)
	return screen
}

func makeCells(width, height int) [][]string {
	cells := make([][]string, height)
	for row := range cells {
		cells[row] = make([]string, width)
	}
	return cells
}

func (s *terminalScreen) consumeEscape(data []byte, index int) int {
	if index+1 >= len(data) {
		return len(data)
	}
	next := data[index+1]
	switch next {
	case '[':
		end := index + 2
		for end < len(data) && (data[end] < 0x40 || data[end] > 0x7e) {
			end++
		}
		if end >= len(data) {
			return len(data)
		}
		s.applyCSI(string(data[index+2:end]), data[end])
		return end + 1
	case ']':
		return consumeStringEscape(data, index+2)
	case 'P', 'X', '^', '_':
		return consumeStringEscape(data, index+2)
	case '7':
		s.savedRow, s.savedCol = s.row, s.col
	case '8':
		s.row, s.col = s.savedRow, s.savedCol
		s.clampCursor()
	case 'D':
		s.lineFeed()
	case 'E':
		s.col = 0
		s.lineFeed()
	case 'M':
		if s.row > s.scrollTop {
			s.row--
		}
	case 'c':
		s.clearAll()
		s.row, s.col = 0, 0
	}
	return index + 2
}

func consumeStringEscape(data []byte, index int) int {
	for index < len(data) {
		if data[index] == 0x07 {
			return index + 1
		}
		if data[index] == 0x1b && index+1 < len(data) && data[index+1] == '\\' {
			return index + 2
		}
		index++
	}
	return len(data)
}

func (s *terminalScreen) applyCSI(body string, final byte) {
	private := strings.HasPrefix(body, "?") || strings.HasPrefix(body, ">")
	clean := strings.TrimLeft(body, "?><!")
	params := parseCSIParams(clean)
	first := csiParam(params, 0, 1)
	if final != 'm' {
		s.wrapPending = false
	}
	switch final {
	case 'A':
		s.row -= first
	case 'B':
		s.row += first
	case 'C', 'a':
		s.col += first
	case 'D':
		s.col -= first
	case 'E':
		s.row += first
		s.col = 0
	case 'F':
		s.row -= first
		s.col = 0
	case 'G', '`':
		s.ensureWidth(first)
		s.col = first - 1
	case 'd':
		s.ensureHeight(first)
		s.row = first - 1
	case 'H', 'f':
		row := csiParam(params, 0, 1)
		col := csiParam(params, 1, 1)
		s.ensureHeight(row)
		s.ensureWidth(col)
		s.row, s.col = row-1, col-1
	case 'J':
		s.eraseDisplay(csiParam(params, 0, 0))
	case 'K':
		s.eraseLine(csiParam(params, 0, 0))
	case 'X':
		s.eraseCells(first)
	case 'P':
		s.deleteCells(first)
	case '@':
		s.insertCells(first)
	case 'L':
		s.insertLines(first)
	case 'M':
		s.deleteLines(first)
	case 'S':
		s.scrollUp(first)
	case 'T':
		s.scrollDown(first)
	case 'r':
		if !private {
			top := csiParam(params, 0, 1) - 1
			bottom := csiParam(params, 1, s.height) - 1
			if top >= 0 && bottom >= top && bottom < s.height {
				s.scrollTop, s.scrollBottom = top, bottom
			}
		}
	case 's':
		s.savedRow, s.savedCol = s.row, s.col
	case 'u':
		s.row, s.col = s.savedRow, s.savedCol
	case 'h':
		if private && containsCSIParam(params, 1049) {
			s.clearAll()
			s.row, s.col = 0, 0
		}
	}
	s.clampCursor()
}

func parseCSIParams(body string) []int {
	if body == "" {
		return nil
	}
	parts := strings.Split(body, ";")
	params := make([]int, len(parts))
	for index, part := range parts {
		part, _, _ = strings.Cut(part, ":")
		if value, err := strconv.Atoi(part); err == nil {
			params[index] = value
		}
	}
	return params
}

func csiParam(params []int, index, fallback int) int {
	if index >= len(params) || params[index] == 0 {
		return fallback
	}
	return params[index]
}

func containsCSIParam(params []int, target int) bool {
	for _, value := range params {
		if value == target {
			return true
		}
	}
	return false
}

func (s *terminalScreen) ensureWidth(width int) {
	if width <= s.width || width > maxTerminalWidth {
		return
	}
	for row := range s.cells {
		s.cells[row] = append(s.cells[row], make([]string, width-s.width)...)
	}
	s.width = width
}

func (s *terminalScreen) ensureHeight(height int) {
	if height <= s.height || height > maxTerminalHeight {
		return
	}
	for len(s.cells) < height {
		s.cells = append(s.cells, make([]string, s.width))
	}
	s.height = height
	s.scrollBottom = height - 1
}

func (s *terminalScreen) clampCursor() {
	s.row = maxInt(0, minInt(s.row, s.height-1))
	s.col = maxInt(0, minInt(s.col, s.width-1))
}

func (s *terminalScreen) writeRune(r rune) {
	if s.wrapPending {
		s.col = 0
		s.lineFeed()
		s.wrapPending = false
	}
	if unicode.Is(unicode.Mn, r) && s.col > 0 {
		s.cells[s.row][s.col-1] += string(r)
		return
	}
	width := terminalRuneWidth(r)
	if s.col+width > s.width {
		s.col = 0
		s.lineFeed()
	}
	s.cells[s.row][s.col] = string(r)
	if width == 2 && s.col+1 < s.width {
		s.cells[s.row][s.col+1] = "\x00"
	}
	s.col += width
	if s.col >= s.width {
		s.col = s.width - 1
		s.wrapPending = true
	}
}

func terminalRuneWidth(r rune) int {
	if r >= 0x1100 && (r <= 0x115f || r == 0x2329 || r == 0x232a ||
		(r >= 0x2e80 && r <= 0xa4cf) || (r >= 0xac00 && r <= 0xd7a3) ||
		(r >= 0xf900 && r <= 0xfaff) || (r >= 0xfe10 && r <= 0xfe6f) ||
		(r >= 0xff00 && r <= 0xff60) || (r >= 0xffe0 && r <= 0xffe6) ||
		(r >= 0x1f300 && r <= 0x1faff) || (r >= 0x20000 && r <= 0x3fffd)) {
		return 2
	}
	return 1
}

func (s *terminalScreen) lineFeed() {
	if s.row == s.scrollBottom {
		s.scrollUp(1)
		return
	}
	if s.row < s.height-1 {
		s.row++
	}
}

func (s *terminalScreen) clearAll() {
	s.cells = makeCells(s.width, s.height)
}

func (s *terminalScreen) eraseDisplay(mode int) {
	switch mode {
	case 2, 3:
		s.clearAll()
	case 1:
		for row := 0; row < s.row; row++ {
			s.cells[row] = make([]string, s.width)
		}
		for col := 0; col <= s.col; col++ {
			s.cells[s.row][col] = ""
		}
	default:
		for col := s.col; col < s.width; col++ {
			s.cells[s.row][col] = ""
		}
		for row := s.row + 1; row < s.height; row++ {
			s.cells[row] = make([]string, s.width)
		}
	}
}

func (s *terminalScreen) eraseLine(mode int) {
	switch mode {
	case 1:
		for col := 0; col <= s.col; col++ {
			s.cells[s.row][col] = ""
		}
	case 2:
		s.cells[s.row] = make([]string, s.width)
	default:
		for col := s.col; col < s.width; col++ {
			s.cells[s.row][col] = ""
		}
	}
}

func (s *terminalScreen) eraseCells(count int) {
	for col := s.col; col < minInt(s.width, s.col+count); col++ {
		s.cells[s.row][col] = ""
	}
}

func (s *terminalScreen) deleteCells(count int) {
	count = minInt(count, s.width-s.col)
	copy(s.cells[s.row][s.col:], s.cells[s.row][s.col+count:])
	for col := s.width - count; col < s.width; col++ {
		s.cells[s.row][col] = ""
	}
}

func (s *terminalScreen) insertCells(count int) {
	count = minInt(count, s.width-s.col)
	copy(s.cells[s.row][s.col+count:], s.cells[s.row][s.col:s.width-count])
	for col := s.col; col < s.col+count; col++ {
		s.cells[s.row][col] = ""
	}
}

func (s *terminalScreen) insertLines(count int) {
	if s.row < s.scrollTop || s.row > s.scrollBottom {
		return
	}
	for ; count > 0; count-- {
		copy(s.cells[s.row+1:s.scrollBottom+1], s.cells[s.row:s.scrollBottom])
		s.cells[s.row] = make([]string, s.width)
	}
}

func (s *terminalScreen) deleteLines(count int) {
	if s.row < s.scrollTop || s.row > s.scrollBottom {
		return
	}
	for ; count > 0; count-- {
		copy(s.cells[s.row:s.scrollBottom], s.cells[s.row+1:s.scrollBottom+1])
		s.cells[s.scrollBottom] = make([]string, s.width)
	}
}

func (s *terminalScreen) scrollUp(count int) {
	for ; count > 0; count-- {
		copy(s.cells[s.scrollTop:s.scrollBottom], s.cells[s.scrollTop+1:s.scrollBottom+1])
		s.cells[s.scrollBottom] = make([]string, s.width)
	}
}

func (s *terminalScreen) scrollDown(count int) {
	for ; count > 0; count-- {
		copy(s.cells[s.scrollTop+1:s.scrollBottom+1], s.cells[s.scrollTop:s.scrollBottom])
		s.cells[s.scrollTop] = make([]string, s.width)
	}
}

func (s *terminalScreen) String() string {
	lines := make([]string, len(s.cells))
	first, last := -1, -1
	for row, cells := range s.cells {
		var line strings.Builder
		for _, cell := range cells {
			if cell == "" {
				line.WriteByte(' ')
			} else if cell != "\x00" {
				line.WriteString(cell)
			}
		}
		lines[row] = strings.TrimRight(line.String(), " ")
		if lines[row] != "" {
			if first < 0 {
				first = row
			}
			last = row
		}
	}
	if first < 0 {
		return ""
	}
	return strings.Join(lines[first:last+1], "\n")
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
