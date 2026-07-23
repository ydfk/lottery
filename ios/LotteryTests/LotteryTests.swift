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
        var draft = TicketDraft(issue: "2026088", entries: [valid])
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
        let draft = TicketDraft(lottery: .dlt, issue: "26088", entries: [entry])
        #expect(entry.isValid(for: .dlt))
        #expect(draft.calculatedCost == 9)
        #expect(draft.recognizedCostDiffers(from: 6))
        #expect(!draft.recognizedCostDiffers(from: 9))
    }

    @Test("多注与不同倍数合计")
    func multipleEntriesCost() {
        let standard = TicketEntryDraft(
            red: [1, 2, 3, 4, 5],
            blue: [6, 7],
            multiple: 2
        )
        let additional = TicketEntryDraft(
            red: [8, 9, 10, 11, 12],
            blue: [1, 2],
            multiple: 3,
            isAdditional: true
        )
        let draft = TicketDraft(lottery: .dlt, entries: [standard, additional])
        #expect(draft.totalMultiple == 5)
        #expect(draft.validEntryCount == 2)
        #expect(draft.calculatedCost == 13)
        #expect(TicketEntryDraft.clampedMultiple(0) == 1)
        #expect(TicketEntryDraft.clampedMultiple(100) == 99)
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

    @Test("开奖走势统计红蓝号码出现次数")
    func drawTrendFrequency() {
        let first = makeDraw(id: "1", red: "01,02,03,04,05,06", blue: "07")
        let second = makeDraw(id: "2", red: "01,08,09,10,11,12", blue: "07")
        let frequency = DrawTrendCalculator.frequencies(for: [first, second])
        #expect(frequency.red[1] == 2)
        #expect(frequency.red[2] == 1)
        #expect(frequency.blue[7] == 2)
        #expect(frequency.blue[8] == nil)
    }

    @Test("历史票据解码生成日期和关联推荐")
    func ticketRecommendationDecoding() throws {
        let json = #"""
        {
          "id":"ticket-1","createdAt":"2026-07-23T08:30:00+08:00","updatedAt":"2026-07-23T09:00:00+08:00",
          "lotteryCode":"ssq","recommendationId":"recommendation-1","issue":"2026088","manualDrawDate":null,
          "source":"manual","recognizedText":"","status":"pending","costAmount":2,"prizeAmount":0,
          "purchasedAt":"2026-07-23T08:35:00+08:00","checkedAt":null,"notes":"","entries":[],"imageUrl":"",
          "drawDate":"2026-07-25T00:00:00+08:00","drawRedNumbers":"","drawBlueNumbers":"",
          "recommendation":{"id":"recommendation-1","issue":"2026088","drawDate":"2026-07-25T00:00:00+08:00",
            "summary":"测试推荐","createdAt":"2026-07-23T08:00:00+08:00","entries":[]}
        }
        """#
        let ticket = try JSONDecoder().decode(Ticket.self, from: Data(json.utf8))
        #expect(ticket.createdAt == "2026-07-23T08:30:00+08:00")
        #expect(ticket.drawDate == "2026-07-25T00:00:00+08:00")
        #expect(ticket.recommendation?.summary == "测试推荐")
    }

    private func makeDraw(id: String, red: String, blue: String) -> DrawResult {
        DrawResult(
            id: id,
            createdAt: "2026-07-23T00:00:00Z",
            updatedAt: "2026-07-23T00:00:00Z",
            lotteryCode: "ssq",
            issue: id,
            drawDate: "2026-07-23T00:00:00Z",
            redNumbers: red,
            blueNumbers: blue,
            saleAmount: 0,
            prizePoolAmount: 0,
            firstPrizeAmount: 0,
            secondPrizeAmount: 0,
            source: "test",
            rawPayload: "{}",
            prizeDetails: []
        )
    }
}
