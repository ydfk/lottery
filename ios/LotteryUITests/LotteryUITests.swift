import XCTest

final class LotteryUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    @MainActor
    func testLoginForm() throws {
        let app = XCUIApplication()
        app.launchArguments.append("UITEST_RESET_SESSION")
        app.launch()

        let username = app.textFields["username"]
        let password = app.secureTextFields["password"]
        let login = app.buttons["login"]
        XCTAssertTrue(username.waitForExistence(timeout: 5))
        XCTAssertTrue(password.exists)
        XCTAssertFalse(login.isEnabled)

        username.tap()
        username.typeText("alice")
        password.tap()
        password.typeText("secret")
        XCTAssertTrue(login.isEnabled)
    }
}

