package postgres

import (
	"context"
	"ticketer/internal/catalog"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ShowSeatRepository struct {
	db *pgxpool.Pool
}

func NewShowSeatRepository(db *pgxpool.Pool) catalog.ShowSeatRepository {
	return &ShowSeatRepository{
		db: db,
	}
}

func (r *ShowSeatRepository) GetByID(id string) (*catalog.ShowSeat, error) {
	query := `SELECT id, show_id, seat_id, status FROM show_seats WHERE id = $1`
	var seat catalog.ShowSeat
	var status string
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&seat.ID,
		&seat.ShowID,
		&seat.SeatID,
		&status,
	)
	if err != nil {
		return nil, err
	}
	seat.Status = catalog.ShowSeatStatus(status)
	return &seat, nil
}

func (r *ShowSeatRepository) GetByIDs(ids []string) ([]catalog.ShowSeat, error) {
	query := `SELECT id, show_id, seat_id, status FROM show_seats WHERE id = ANY($1)`
	rows, err := r.db.Query(context.Background(), query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []catalog.ShowSeat
	for rows.Next() {
		var seat catalog.ShowSeat
		var status string
		if err := rows.Scan(
			&seat.ID,
			&seat.ShowID,
			&seat.SeatID,
			&status,
		); err != nil {
			return nil, err
		}
		seat.Status = catalog.ShowSeatStatus(status)
		seats = append(seats, seat)
	}
	return seats, nil
}

func (r *ShowSeatRepository) GetByShow(showID string) ([]catalog.ShowSeat, error) {
	query := `SELECT id, show_id, seat_id, status FROM show_seats WHERE show_id = $1`
	rows, err := r.db.Query(context.Background(), query, showID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []catalog.ShowSeat
	for rows.Next() {
		var seat catalog.ShowSeat
		var status string
		if err := rows.Scan(
			&seat.ID,
			&seat.ShowID,
			&seat.SeatID,
			&status,
		); err != nil {
			return nil, err
		}
		seat.Status = catalog.ShowSeatStatus(status)
		seats = append(seats, seat)
	}
	return seats, nil
}

func (r *ShowSeatRepository) Save(showSeat *catalog.ShowSeat) error {
	query := `INSERT INTO show_seats (id, show_id, seat_id, status) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(context.Background(), query,
		showSeat.ID,
		showSeat.ShowID,
		showSeat.SeatID,
		string(showSeat.Status),
	)
	return err
}

func (r *ShowSeatRepository) UpdateStatuses(ids []string, status catalog.ShowSeatStatus) error {
	query := `UPDATE show_seats SET status = $1 WHERE id = ANY($2)`
	_, err := r.db.Exec(context.Background(), query, string(status), ids)
	return err
}

func (r *ShowSeatRepository) GetAvailableSeats(showID string) ([]catalog.ShowSeat, error) {
	query := `SELECT id, show_id, seat_id, status FROM show_seats WHERE show_id = $1 AND status = $2`
	rows, err := r.db.Query(context.Background(), query, showID, string(catalog.ShowSeatStatusAvailable))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []catalog.ShowSeat
	for rows.Next() {
		var seat catalog.ShowSeat
		var status string
		if err := rows.Scan(
			&seat.ID,
			&seat.ShowID,
			&seat.SeatID,
			&status,
		); err != nil {
			return nil, err
		}
		seat.Status = catalog.ShowSeatStatus(status)
		seats = append(seats, seat)
	}
	return seats, nil
}
