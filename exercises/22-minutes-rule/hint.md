규칙은 `ALTER TABLE 표 ADD CONSTRAINT 이름 CHECK (조건)`으로 겁니다.
---
'0보다 크다'에 0은 들어가지 않습니다. 비어 있는(NULL) TV 프로그램은 CHECK가 막지 않습니다.
---
끝에 `ALTER TABLE`이 나오면 터미널 2에서 0분짜리를 INSERT해 보세요. 규칙 이름이 담긴 에러가 나야 합니다.
