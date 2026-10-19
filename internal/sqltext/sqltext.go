// Package sqltext는 학생이 붙여넣은 SQL을 문장으로 나누고, 채점기가 받지 않을 문장을 걸러 낸다.
package sqltext

import "strings"

// Normalize는 Windows 줄바꿈(CRLF)을 LF로 바꾸고 앞뒤 공백을 지운다.
func Normalize(s string) string { return strings.TrimSpace(strings.ReplaceAll(s, "\r\n", "\n")) }

const (
	plain = iota
	single
	double
	lineComment
	blockComment
	dollar
)

type piece struct {
	raw  string // 원문(주석 포함)
	code string // 주석을 공백으로 바꾼 것
}

type scanned struct {
	pieces []piece
	meta   bool // 따옴표·주석 밖에 \가 있다 — psql은 줄 어디에 있든 메타 명령으로 처리한다
	open   int  // 끝났을 때 상태: 닫히지 않은 따옴표·주석이면 plain이 아니다
}

func isIdent(r rune) bool {
	return r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r > 127
}

// dollarTag는 rs[i]가 '$'일 때 $태그$ 모양이면 그 태그 전체("$tag$" 또는 "$$")를 돌려준다. $1 같은 매개변수는 아니다.
func dollarTag(rs []rune, i int) string {
	j := i + 1
	if j < len(rs) && rs[j] >= '0' && rs[j] <= '9' {
		return ""
	}
	for j < len(rs) && rs[j] != '$' && isIdent(rs[j]) {
		j++
	}
	if j < len(rs) && rs[j] == '$' {
		return string(rs[i : j+1])
	}
	return ""
}

// scan은 따옴표(' " $$ E'')·주석을 지키며 세미콜론에서 문장을 끊는다.
func scan(src string) scanned {
	rs := []rune(Normalize(src))
	var r scanned
	var raw, code strings.Builder
	state, escapes, tag := plain, false, ""
	flush := func() {
		if strings.TrimSpace(code.String()) != "" {
			r.pieces = append(r.pieces, piece{strings.TrimSpace(raw.String()), code.String()})
		}
		raw.Reset()
		code.Reset()
	}
	emit := func(s string, comment bool) {
		raw.WriteString(s)
		if comment {
			code.WriteString(strings.Repeat(" ", len([]rune(s))))
		} else {
			code.WriteString(s)
		}
	}
	for i := 0; i < len(rs); i++ {
		c := rs[i]
		var next rune
		if i+1 < len(rs) {
			next = rs[i+1]
		}
		switch state {
		case plain:
			switch {
			case c == ';':
				flush()
				continue
			case c == '\\':
				r.meta = true
			case c == '\'':
				state = single
				escapes = i > 0 && (rs[i-1] == 'E' || rs[i-1] == 'e') && (i < 2 || !isIdent(rs[i-2]))
			case c == '"':
				state = double
			case c == '-' && next == '-':
				state = lineComment
				emit("-", true)
				continue
			case c == '/' && next == '*':
				state = blockComment
				emit("/", true)
				continue
			case c == '$':
				if t := dollarTag(rs, i); t != "" {
					state, tag = dollar, t
					emit(t, false)
					i += len([]rune(t)) - 1
					continue
				}
			}
			emit(string(c), false)
		case single:
			switch {
			case escapes && c == '\\' && next != 0:
				emit(string(c)+string(next), false)
				i++
			case c == '\'' && next == '\'':
				emit("''", false)
				i++
			case c == '\'':
				state = plain
				emit("'", false)
			default:
				emit(string(c), false)
			}
		case double:
			if c == '"' {
				state = plain
			}
			emit(string(c), false)
		case lineComment:
			if c == '\n' {
				state = plain
				emit("\n", false)
				continue
			}
			emit(string(c), true)
		case blockComment:
			if c == '*' && next == '/' {
				emit("*/", true)
				i++
				state = plain
				continue
			}
			emit(string(c), true)
		case dollar:
			if c == '$' && strings.HasPrefix(string(rs[i:]), tag) {
				emit(tag, false)
				i += len([]rune(tag)) - 1
				state = plain
				continue
			}
			emit(string(c), false)
		}
	}
	flush()
	if state != lineComment {
		r.open = state
	}
	return r
}

// Split은 문장 원문들을 돌려준다(끝 세미콜론 없이, 주석만 있는 문장은 버린다).
func Split(sql string) []string {
	var out []string
	for _, p := range scan(sql).pieces {
		out = append(out, p.raw)
	}
	return out
}

func firstWordOfCode(code string) string {
	f := strings.Fields(code)
	if len(f) == 0 {
		return ""
	}
	w := strings.ToUpper(f[0])
	if i := strings.IndexAny(w, "(;"); i >= 0 {
		w = w[:i]
	}
	return w
}

// FirstWord는 앞 주석을 건너뛴 첫 단어를 대문자로 돌려준다.
func FirstWord(stmt string) string {
	ps := scan(stmt).pieces
	if len(ps) == 0 {
		return ""
	}
	return firstWordOfCode(ps[0].code)
}

// Forbidden은 채점기가 실행하지 않을 이유를 한국어로 돌려준다. 괜찮으면 "".
func Forbidden(sql string) string {
	sc := scan(sql)
	switch sc.open {
	case single:
		return "작은따옴표(')가 닫히지 않았습니다. 값의 끝에 '를 붙이세요"
	case double:
		return "큰따옴표(\")가 닫히지 않았습니다. 글자 값은 작은따옴표(')로 감쌉니다"
	case blockComment:
		return "주석(/*)이 닫히지 않았습니다. */로 닫으세요"
	case dollar:
		return "$$로 시작한 글자가 닫히지 않았습니다"
	}
	ps := sc.pieces
	if sc.meta {
		return "정답 칸에는 SQL만 넣습니다. \\로 시작하는 psql 명령은 터미널 2에서 쓰세요"
	}
	for _, p := range ps {
		w := firstWordOfCode(p.code)
		switch w {
		case "BEGIN", "START", "COMMIT", "END", "ROLLBACK", "ABORT", "SAVEPOINT", "RELEASE", "SET", "RESET", "PREPARE":
			return "채점기가 트랜잭션·설정을 직접 다룹니다. " + w + " 문장은 빼고 붙여넣으세요"
		}
		f := strings.Fields(strings.ToUpper(p.code))
		if len(f) >= 2 && (f[0] == "DROP" || f[0] == "CREATE") && (f[1] == "DATABASE" || f[1] == "SCHEMA") {
			return "DB·스키마를 지우거나 만드는 문장은 채점하지 않습니다"
		}
	}
	return ""
}
