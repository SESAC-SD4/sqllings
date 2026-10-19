package console

// Init은 Windows에서 콘솔 출력 코드페이지를 UTF-8(65001)로 설정한다.
// 다른 플랫폼에서는 아무것도 하지 않는다.
func Init() {
	initPlatform()
}
