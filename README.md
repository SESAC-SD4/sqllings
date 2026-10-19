# sqllings — SQL 연습 문제

진짜 PostgreSQL에서 **진짜 psql로 시험해 보고**, 된 SQL을 **붙여넣으면 자동 채점**되는 한국어 SQL 연습 문제입니다. 수업 날마다 그날 배운 것을 연습하는 장이 하나씩 열립니다 — 새 장이 열리면 `git pull`.

## 준비

- PostgreSQL 18과 `psql` (수업 1일차에 설치)
- 사용자 `sd4`와 연습용 DB `sqllings` — 처음 한 번, **관리자로 접속해** 만듭니다
  ```bash
  psql -d postgres          # mac
  psql -U postgres          # Windows (Git Bash, 비밀번호 postgres)
  ```
  ```sql
  CREATE DATABASE sqllings OWNER sd4;
  \q
  ```
  - `CREATE DATABASE`가 나오면 끝. 이미 만들었으면 `already exists`(한국어 서버: `이미 있음`)가 나옵니다 — 그대로 두면 됩니다
  - 만들지 않고 `./sqllings`를 켜면 같은 안내가 화면에 나옵니다

## 터미널 두 개

| | 터미널 1 — 문제 화면 | 터미널 2 — 시험하는 곳 |
| --- | --- | --- |
| 실행 | `./sqllings` (Windows: `./sqllings.exe`) | `psql -U sd4 -d sqllings` |
| 하는 일 | 문제 보기 · 정답 붙여넣기 · 채점 · 힌트 | SQL을 마음껏 쳐 본다 |

```bash
cd ~
git clone <강사가 알려 준 주소> sqllings
cd sqllings
./sqllings          # mac
./sqllings.exe      # Windows (Git Bash)
```

> Windows에서 **Defender 경고**가 뜨면: "추가 정보" → "그래도 실행"

> Windows의 터미널 2에서 `UTF8 … UHC 인코딩에는 없습니다` 오류가 나면 **SQL이 틀린 게 아닙니다.** `é` 같은 글자가 든 행(전체의 약 24%)을 Windows 콘솔이 표시하지 못해서입니다. 그 SQL을 그대로 터미널 1에 붙여 보세요. 채점은 정상입니다.

## 한 문제 푸는 흐름

1. 터미널 1에서 문제를 고른다 (`↑` `↓`)
2. 터미널 2의 psql에서 SQL을 쳐 본다. 결과를 눈으로 확인한다
3. 된 SQL을 복사해서 터미널 1의 **정답 칸에 붙여넣고 Enter**
4. ✅이면 `↓`로 다음 문제. ❌이면 안내와 결과 화면을 읽고, 막히면 **Tab**(힌트)

채점은 늘 **그 장의 처음 데이터**로 합니다. 터미널 2에서 데이터를 망가뜨려도 채점에는 영향이 없고, **Ctrl+R**이면 터미널 2의 데이터도 처음으로 돌아갑니다.

| 키 | 하는 일 |
| --- | --- |
| `↑` `↓` | 문제 고르기 |
| **Enter** | 채점 |
| **Tab** | 힌트 (누를 때마다 한 단계, 3단계까지) · **Esc**로 닫기 |
| Ctrl+L | 정답 칸 비우기 |
| **Ctrl+R** | 터미널 2의 데이터(`public`)를 이 문제의 장 처음 상태로 (`y`로 확인) |
| Ctrl+X | ✅ 기록 전체 지우기 (`y`로 확인) |
| PgUp / PgDn | 문제 설명 스크롤 (mac 노트북: `fn` + `↑` `↓`) |
| **Ctrl+C** | 끝내기 |

정답 칸에는 **SQL만** 넣습니다. `\d` 같은 psql 명령은 터미널 2에서 쓰세요.

## 화면이 안 뜰 때 (명령으로 풀기)

```bash
./sqllings list              # 목록
./sqllings start 7           # 7번 문제 보기
./sqllings check 7 답.sql    # 파일에 적은 SQL로 채점
./sqllings hint 7            # 힌트
./sqllings reset 7           # 터미널 2의 데이터를 7번 문제의 장 처음 상태로
```

## 데이터 출처

*Netflix Movies and TV Shows* — Shivam Bansal, Kaggle (`shivamb/netflix-shows`), **CC0: Public Domain**

## 라이선스

MIT. 화면 구조는 같은 수업의 shellings·gitlings와 같습니다.
