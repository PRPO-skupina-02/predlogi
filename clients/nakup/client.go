package nakup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type ReservationType string

const (
	Online ReservationType = "ONLINE"
	Pos    ReservationType = "POS"
)

type Reservation struct {
	ID         uuid.UUID       `json:"id"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
	TimeSlotID uuid.UUID       `json:"timeslot_id"`
	UserID     uuid.UUID       `json:"user_id"`
	Type       ReservationType `json:"type"`
	Row        int             `json:"row"`
	Col        int             `json:"col"`
}

type ReservationsResponse struct {
	Data []Reservation `json:"data"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetUserReservations(userID uuid.UUID) ([]Reservation, error) {
	url := fmt.Sprintf("%s/api/v1/nakup/reservations?user_id=%s", c.baseURL, userID.String())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	var reservationsResp ReservationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&reservationsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return reservationsResp.Data, nil
}
