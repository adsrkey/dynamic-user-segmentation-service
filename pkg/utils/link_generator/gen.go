package linkgenerator

import (
	"errors"
	"os"
	"syscall"
	"time"

	dto "github.com/adsrkey/dynamic-user-segmentation-service/internal/dto/handler/user"
	"github.com/adsrkey/dynamic-user-segmentation-service/pkg/csv"
)

const dir = "./files/"

func GenerateReportsLink(reports []dto.Report, host string) (string, error) {
	data := make([][]string, 0, len(reports))

	for _, report := range reports {
		data = append(data, []string{report.ID.String(), report.Segment, report.Operation, report.OperationAt.Format(time.RFC822)})
	}

	err := os.Mkdir(dir, os.FileMode(0755))
	if err != nil {
		if errors.Is(err, syscall.EEXIST) {
		} else {
			return "", err
		}
	}

	fileId, err := csv.CreateFile(data, dir)

	if err != nil {
		return "", err
	}

	return host + "/api/v1/user/files?file_id=" + fileId.String(), nil
}
