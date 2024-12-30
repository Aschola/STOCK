package utils
import "time"

func parseExpiryDate(dateStr string) (*time.Time, error) {
    // Define the format that matches your input
    const layout = "2006-01-02T15:04:05Z07:00"
    
    parsedTime, err := time.Parse(layout, dateStr)
    if err != nil {
        return nil, err
    }
    return &parsedTime, nil
}
