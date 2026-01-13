package spored

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Movie struct {
	ID            uuid.UUID `json:"id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	ImageURL      string    `json:"image_url"`
	Rating        float64   `json:"rating"`
	LengthMinutes int       `json:"length_minutes"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type TimeSlot struct {
	ID        uuid.UUID `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	RoomID    uuid.UUID `json:"room_id"`
	MovieID   uuid.UUID `json:"movie_id"`
	Movie     Movie     `json:"movie"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TimeSlotsResponse struct {
	Data []TimeSlot `json:"data"`
}

type MovieResponse struct {
	Data Movie `json:"data"`
}

type MoviesResponse struct {
	Data []Movie `json:"data"`
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

func (c *Client) GetTimeSlot(timeSlotID uuid.UUID) (*TimeSlot, error) {
	url := fmt.Sprintf("%s/api/v1/spored/timeslots/%s", c.baseURL, timeSlotID.String())

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

	var timeSlot TimeSlot
	if err := json.NewDecoder(resp.Body).Decode(&timeSlot); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &timeSlot, nil
}

func (c *Client) GetUpcomingTimeSlots(startDate, endDate time.Time) ([]TimeSlot, error) {
	url := fmt.Sprintf("%s/api/v1/spored/timeslots?start_date=%s&end_date=%s",
		c.baseURL,
		startDate.Format(time.RFC3339),
		endDate.Format(time.RFC3339),
	)

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

	var timeSlotsResp TimeSlotsResponse
	if err := json.NewDecoder(resp.Body).Decode(&timeSlotsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return timeSlotsResp.Data, nil
}

func (c *Client) GetMovie(movieID uuid.UUID) (*Movie, error) {
	url := fmt.Sprintf("%s/api/v1/spored/movies/%s", c.baseURL, movieID.String())

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

	var movieResp MovieResponse
	if err := json.NewDecoder(resp.Body).Decode(&movieResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &movieResp.Data, nil
}

func (c *Client) GetActiveMovies() ([]Movie, error) {
	url := fmt.Sprintf("%s/api/v1/spored/movies?active=true", c.baseURL)

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

	var moviesResp MoviesResponse
	if err := json.NewDecoder(resp.Body).Decode(&moviesResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return moviesResp.Data, nil
}
