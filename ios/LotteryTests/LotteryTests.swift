import Foundation
import Testing
@testable import Lottery

struct LotteryTests {
    @Test("双色球选号规则与金额计算")
    func ticketValidationAndCost() {
        let valid = TicketEntryDraft(
            red: [1, 2, 3, 4, 5, 6],
            blue: [7],
            multiple: 2
        )
        var draft = TicketDraft(issue: "2026088", costAmount: 4, entries: [valid])
        #expect(valid.isValid(for: .ssq))
        #expect(draft.calculatedCost == 4)
        #expect(draft.isValid)

        draft.lottery = .dlt
        #expect(!draft.entries[0].isValid(for: .dlt))
    }

    @Test("大乐透追加金额计算")
    func additionalCost() {
        let entry = TicketEntryDraft(
            red: [1, 2, 3, 4, 5],
            blue: [6, 7],
            multiple: 3,
            isAdditional: true
        )
        let draft = TicketDraft(lottery: .dlt, issue: "26088", costAmount: 9, entries: [entry])
        #expect(entry.isValid(for: .dlt))
        #expect(draft.calculatedCost == 9)
    }

    @Test("后端统一响应可以解码")
    func envelopeDecoding() throws {
        let json = #"{"flag":true,"code":200,"data":{"token":"abc"},"time":"2026-07-23T00:00:00Z"}"#
        let envelope = try JSONDecoder().decode(APIEnvelope<AuthToken>.self, from: Data(json.utf8))
        #expect(envelope.flag)
        #expect(envelope.data?.token == "abc")
    }

    @Test("日期兼容纳秒与普通 RFC3339")
    func dateParsing() {
        #expect(DateParser.date("2026-07-23T08:00:00Z") != nil)
        #expect(DateParser.date("2026-07-23T08:00:00.123456789Z") != nil)
        #expect(DateParser.date("2026-07-23") != nil)
    }

    @Test("号码格式固定为两位")
    func numberFormatting() {
        #expect([1, 9, 12].formattedNumbers == "01,09,12")
        #expect("01,09,12".lotteryNumbers == [1, 9, 12])
    }
}

