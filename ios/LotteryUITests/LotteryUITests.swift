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

        XCTAssertFalse(app.staticTexts["分彩种统计"].exists)
        XCTAssertFalse(app.staticTexts["最新开奖"].exists)
        XCTAssertFalse(app.staticTexts["最新推荐"].exists)
        XCTAssertFalse(app.staticTexts["最近记录"].exists)
        app.tabBars.buttons["推荐"].tap()
        XCTAssertTrue(app.buttons["generate-recommendation"].waitForExistence(timeout: 3))

        let drawsTab = app.tabBars.buttons["开奖"]
        XCTAssertTrue(drawsTab.waitForExistence(timeout: 5))
        XCTAssertEqual(app.tabBars.buttons.count, 4)
        XCTAssertFalse(app.tabBars.buttons["记录"].exists)
        drawsTab.tap()
        XCTAssertTrue(app.buttons["号码走势"].waitForExistence(timeout: 3))
        XCTAssertTrue(app.buttons["开奖列表"].exists)
        XCTAssertTrue(app.navigationBars["开奖"].buttons["筛选"].exists)
        XCTAssertTrue(app.navigationBars["开奖"].buttons["同步"].exists)
        app.buttons["号码走势"].tap()
        XCTAssertFalse(app.navigationBars["开奖"].buttons["筛选"].exists)
        app.buttons["开奖列表"].tap()

        app.tabBars.buttons["历史"].tap()
        XCTAssertFalse(app.navigationBars["历史"].buttons["导入"].exists)
        let newTicket = app.buttons["new-ticket"]
        XCTAssertTrue(newTicket.waitForExistence(timeout: 3))
        newTicket.tap()
        let increaseMultiple = app.buttons["增加倍数"]
        XCTAssertTrue(increaseMultiple.waitForExistence(timeout: 3))
        increaseMultiple.tap()
        increaseMultiple.tap()
        increaseMultiple.tap()
        XCTAssertTrue(app.staticTexts["ticket-multiple"].waitForExistence(timeout: 2))
        XCTAssertEqual(app.staticTexts["ticket-multiple"].label, "4 倍")
        XCTAssertTrue(app.descendants(matching: .any)["ticket-cost"].exists)

        let issue = app.textFields["ticket-issue"]
        issue.tap()
        issue.typeText("2026088")
        app.buttons["close-ticket-editor"].tap()
        XCTAssertTrue(app.buttons["丢弃并关闭"].waitForExistence(timeout: 2))
        app.buttons["丢弃并关闭"].tap()

        XCTAssertTrue(newTicket.waitForExistence(timeout: 3))
        newTicket.tap()
        XCTAssertEqual(app.textFields["ticket-issue"].value as? String, "期号")
    }
}
