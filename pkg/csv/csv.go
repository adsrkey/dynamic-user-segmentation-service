package csv

import (
	"encoding/csv"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func CreateFile(data [][]string, dir string) (uuid.UUID, error) {
	id := uuid.New()

	filePath, _ := filepath.Abs(dir + id.String() + ".csv")

	file, err := os.Create(filePath)
	if err != nil {
		return uuid.UUID{}, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = ';'

	err = writer.WriteAll(data)
	if err != nil {
		return uuid.UUID{}, err
	}

	return id, nil
}
