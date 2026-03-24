package lottery

import (
	"io"
	"mime/multipart"

	"go-fiber-starter/internal/api/response"
	lotteryService "go-fiber-starter/internal/service/lottery"

	"github.com/gofiber/fiber/v2"
)

// @Summary 批量导入历史票据
// @Description 上传 Excel 批量导入票据，图片压缩包可选；不依赖 OCR，可直接按表格号码入库
// @Tags lottery
// @Accept mpfd
// @Produce json
// @Security BearerAuth
// @Param workbook formData file true "Excel 文件，支持 xlsx"
// @Param imagesZip formData file false "图片压缩包，Excel 中 imageName 列会按文件名匹配"
// @Success 200 {object} TicketImportResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /lotteries/tickets/import [post]
func ImportTickets(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}

	workbookFile, err := c.FormFile("workbook")
	if err != nil {
		return response.Error(c, "请上传 Excel 文件", fiber.StatusBadRequest)
	}

	workbookData, err := readUploadedFile(workbookFile)
	if err != nil {
		return err
	}

	imagesArchive := []byte(nil)
	if archiveFile, archiveErr := c.FormFile("imagesZip"); archiveErr == nil {
		imagesArchive, err = readUploadedFile(archiveFile)
		if err != nil {
			return err
		}
	}

	data, err := lotteryService.ImportTickets(c.Context(), lotteryService.ImportTicketsInput{
		UserID:        userID,
		Workbook:      workbookData,
		ImagesArchive: imagesArchive,
	})
	if err != nil {
		return err
	}
	return response.Success(c, data)
}

func readUploadedFile(fileHeader *multipart.FileHeader) ([]byte, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}
