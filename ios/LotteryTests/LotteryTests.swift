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

    @Test("号码命中区分红球和蓝球")
    func numberMatch() {
        let match = LotteryNumberMatch(
            redNumbers: "01,02,03,04,05,06",
            blueNumbers: "07",
            drawRedNumbers: "01,03,08,09,10,11",
            drawBlueNumbers: "07"
        )
        #expect(match.hasDrawNumbers)
        #expect(match.redHitCount == 2)
        #expect(match.blueHitCount == 1)
        #expect(match.isRedHit(1))
        #expect(!match.isRedHit(2))

        let pending = LotteryNumberMatch(
            redNumbers: "01,02,03,04,05,06",
            blueNumbers: "07",
            drawRedNumbers: "",
            drawBlueNumbers: ""
        )
        #expect(!pending.hasDrawNumbers)
    }

    @Test("推荐和历史列表完整映射多注号码")
    func listEntryDisplayMapping() {
        let recommendations = LotteryEntryDisplay.recommendations([
            recommendationEntry(id: "r1", sequence: 1),
            recommendationEntry(id: "r2", sequence: 2),
        ])
        #expect(recommendations.map(\.sequence) == [1, 2])
        #expect(recommendations.allSatisfy { $0.detail == nil })

        let tickets = LotteryEntryDisplay.tickets([
            ticketEntry(id: "t1", sequence: 1, multiple: 2, additional: false),
            ticketEntry(id: "t2", sequence: 2, multiple: 3, additional: true),
        ])
        #expect(tickets.count == 2)
        #expect(tickets[0].detail == "2 倍")
        #expect(tickets[1].detail == "3 倍 · 追加")
        #expect(LotteryEntryDisplay.tickets([]).isEmpty)
    }

    @Test("看板比率和趋势汇总处理零值")
    func dashboardMetrics() {
        let stats = DashboardStats(
            totalTickets: 4,
            wonTickets: 1,
            totalCost: 100,
            totalPrize: 25,
            totalRecommendations: 5,
            purchasedRecommendations: 2
        )
        #expect(stats.winningRate == 0.25)
        #expect(stats.returnRate == 0.25)
        #expect(stats.recommendationPurchaseRate == 0.4)

        let empty = DashboardStats(
            totalTickets: 0,
            wonTickets: 0,
            totalCost: 0,
            totalPrize: 0,
            totalRecommendations: 0,
            purchasedRecommendations: 0
        )
        #expect(empty.winningRate == 0)
        #expect(empty.returnRate == 0)
        #expect(empty.recommendationPurchaseRate == 0)

        let trend = DashboardTrend(
            range: "30d",
            bucket: "day",
            items: [
                DashboardTrendPoint(
                    periodStart: "2026-07-23",
                    totalCost: 10,
                    totalPrize: 20,
                    ticketCount: 2,
                    wonTickets: 1,
                    recommendationCount: 3
                ),
            ]
        )
        #expect(trend.totalCost == 10)
        #expect(trend.totalPrize == 20)
        #expect(trend.ticketCount == 2)
        #expect(trend.recommendationCount == 3)
    }

    @Test("录票入口清空旧内容并保留业务预填")
    @MainActor
    func editorDraftLifecycle() throws {
        let session = AppSession()
        session.ticketDraft.issue = "旧草稿"
        session.startNewTicket()
        #expect(session.ticketDraft.issue.isEmpty)
        #expect(session.isEditorPresented)

        let recommendationJSON = #"""
        {
          "id":"recommendation-1","lotteryCode":"ssq","issue":"2026088","drawDate":"2026-07-25",
          "provider":"test","model":"test","strategy":"test","promptVersion":"v1","summary":"测试",
          "basis":"测试","checkedAt":null,"prizeAmount":0,"createdAt":"2026-07-23T08:00:00+08:00",
          "entries":[{"id":"entry-1","sequence":1,"redNumbers":"01,02,03,04,05,06","blueNumbers":"07",
            "confidence":0.8,"reason":"测试","isWinning":false,"prizeName":"","prizeAmount":0,"matchSummary":""}]
        }
        """#
        let recommendation = try JSONDecoder().decode(
            Recommendation.self,
            from: Data(recommendationJSON.utf8)
        )
        let quickSSQ = TicketDraft.purchase(recommendation, multiple: 3, additional: true)
        #expect(quickSSQ.calculatedCost == 6)
        #expect(quickSSQ.entries.first?.isAdditional == false)
        #expect(quickSSQ.entries.first?.payload().redNumbers == "01,02,03,04,05,06")
        #expect(quickSSQ.recommendationId == recommendation.id)
        #expect(quickSSQ.issue == recommendation.issue)

        let dltJSON = recommendationJSON
            .replacingOccurrences(of: "ssq", with: "dlt")
            .replacingOccurrences(of: "01,02,03,04,05,06", with: "01,02,03,04,05")
            .replacingOccurrences(of: "\"blueNumbers\":\"07\"", with: "\"blueNumbers\":\"07,08\"")
        let dlt = try JSONDecoder().decode(Recommendation.self, from: Data(dltJSON.utf8))
        let quickDLT = TicketDraft.purchase(dlt, multiple: 4, additional: true)
        #expect(quickDLT.isValid)
        #expect(quickDLT.calculatedCost == 12)
        #expect(quickDLT.entries.first?.payload().blueNumbers == "07,08")
        #expect(TicketDraft.purchase(dlt, multiple: 4).calculatedCost == 8)
        #expect(TicketDraft.purchase(dlt, multiple: 100).entries.first?.multiple == 99)
        #expect(TicketDraft.purchase(dlt, multiple: 0).entries.first?.multiple == 1)

        session.recordPurchase(recommendation)
        #expect(session.ticketDraft.issue == "2026088")
        #expect(session.ticketDraft.recommendationId == "recommendation-1")
        #expect(session.ticketDraft.entries.first?.red == [1, 2, 3, 4, 5, 6])

        session.closeEditor(saved: false)
        #expect(session.ticketDraft.issue.isEmpty)
        #expect(!session.isEditorPresented)
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

    private func recommendationEntry(id: String, sequence: Int) -> RecommendationEntry {
        RecommendationEntry(
            id: id,
            sequence: sequence,
            redNumbers: "01,02,03,04,05,06",
            blueNumbers: "07",
            confidence: 0.8,
            reason: "",
            isWinning: false,
            prizeName: "",
            prizeAmount: 0,
            matchSummary: ""
        )
    }

    private func ticketEntry(
        id: String,
        sequence: Int,
        multiple: Int,
        additional: Bool
    ) -> TicketEntry {
        TicketEntry(
            id: id,
            sequence: sequence,
            redNumbers: "01,02,03,04,05,06",
            blueNumbers: "07",
            multiple: multiple,
            isAdditional: additional,
            isWinning: false,
            prizeName: "",
            prizeAmount: 0,
            matchSummary: ""
        )
    }

}
