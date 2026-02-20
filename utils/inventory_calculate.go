package utils

import (
	"log/slog"
)

func ProcessInventory(currentQuantity, change int32, transactionType string) int32 {
	var finalQuantity int32
	switch transactionType {
	case "IMPORT":
		finalQuantity = currentQuantity + change
	case "EXPORT":
		if change > currentQuantity {
			slog.Error("Error in calculating final quantity", slog.Any("error", "export must be less than current quantity"))
		}
		finalQuantity = currentQuantity - change
	}
	return finalQuantity
}
