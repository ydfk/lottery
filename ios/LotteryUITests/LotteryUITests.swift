import XCTest

final class LotteryUITests: XCTestCase {
    override func setUpWithError() throws {
        continueAfterFailure = false
    }

    @MainActor
    func testLoginForm() throws {
        let app = XCUIApplication()
        app.launchEnvironment["UITEST_RESET_SESSION"] = "1"
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

    @MainActor
    func testMainTabsAndDrawViews() throws {
        let app = XCUIApplication()
        app.launchEnvironment["UITEST_RESET_SESSION"] = "1"
        app.launchEnvironment["UITEST_SIGNED_IN"] = "1"
        app.launch()

        let drawsTab = app.tabBars.buttons["开奖"]
        XCTAssertTrue(drawsTab.waitForExistence(timeout: 5))
        drawsTab.tap()
        XCTAssertTrue(app.buttons["号码走势"].waitForExistence(timeout: 3))
        XCTAssertTrue(app.buttons["开奖列表"].exists)

        app.buttons["开奖列表"].tap()
        XCTAssertTrue(app.navigationBars["开奖"].buttons["筛选"].exists)
        XCTAssertTrue(app.navigationBars["开奖"].buttons["同步"].exists)

        app.tabBars.buttons["历史"].tap()
        XCTAssertFalse(app.navigationBars["历史"].buttons["导入"].exists)

        app.tabBars.buttons["记录"].tap()
        XCTAssertTrue(app.buttons["增加倍数"].waitForExistence(timeout: 3))
        XCTAssertTrue(app.descendants(matching: .any)["ticket-cost"].exists)
    }
}
