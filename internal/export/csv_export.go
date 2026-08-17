package export

import (
	"encoding/csv"
	"io"

	"maskreview/internal/domain"
)

func WritePendingAuditCSV(w io.Writer, requests []*domain.ChangeRequest) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()
	if err := cw.Write([]string{"ID", "MaskVersionID", "Status", "CreatedAt"}); err != nil {
		return err
	}
	for _, cr := range requests {
		if err := cw.Write([]string{cr.ID, cr.MaskVersionID, string(cr.Status), cr.CreatedAt.Format("2006-01-02T15:04:05Z07:00")}); err != nil {
			return err
		}
	}
	return nil
}
