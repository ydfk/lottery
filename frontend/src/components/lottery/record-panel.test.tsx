import { render, screen } from "@testing-library/react";
import { RecordPanel } from "./record-panel";

test("规则金额只读并说明由服务端确认", () => {
  render(
    <RecordPanel
      displayMode="app"
      selectedRecommendation={null}
      previewUrl=""
      selectedImage={null}
      uploadPending={false}
      uploadedTicket={null}
      recognitionDraft={null}
      lotteryCode="ssq"
      recognizePending={false}
      issue="2026088"
      drawDate="2026-07-23"
      purchasedAt="2026-07-23T12:00"
      costAmount="4.00"
      notes=""
      entryDrafts={[
        {
          redNumbers: "01 02 03 04 05 06",
          blueNumbers: "07",
          multiple: 2,
          isAdditional: false,
        },
      ]}
      submitPending={false}
      onSelectImage={vi.fn()}
      onLotteryCodeChange={vi.fn()}
      onRecognize={vi.fn()}
      onIssueChange={vi.fn()}
      onDrawDateChange={vi.fn()}
      onPurchasedAtChange={vi.fn()}
      onNotesChange={vi.fn()}
      onEntryFieldChange={vi.fn()}
      onToggleEntryAdditional={vi.fn()}
      onChangeEntryMultiple={vi.fn()}
      onAddEntry={vi.fn()}
      onRemoveEntry={vi.fn()}
      onCreateTicket={vi.fn()}
      onClearRecommendation={vi.fn()}
    />
  );

  expect(screen.getByLabelText("规则计算金额")).toHaveAttribute("readonly");
  expect(screen.getByText(/保存后由服务端再次确认/)).toBeInTheDocument();
});
