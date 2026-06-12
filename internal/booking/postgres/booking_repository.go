package postgres

import (
	"context"
	"ticketer/internal/booking"

	"github.com/jackc/pgx/v5/pgxpool"
)

type BookingRepository struct {
	db *pgxpool.Pool
}

func NewBookingRepository(db *pgxpool.Pool) booking.BookingRepository {
	return &BookingRepository{
		db: db,
	}
}

func (r *BookingRepository) Save(b *booking.Booking) (*booking.Booking, error) {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	query := `INSERT INTO bookings (id, user_id, show_id, price, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = tx.Exec(context.Background(), query,
		b.ID,
		b.UserID,
		b.ShowID,
		b.Price,
		string(b.Status),
		b.CreatedAt,
		b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	for _, seatID := range b.SeatIDs {
		seatQuery := `INSERT INTO booking_seats (booking_id, seat_id) VALUES ($1, $2)`
		_, err = tx.Exec(context.Background(), seatQuery, b.ID, seatID)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r *BookingRepository) UpdateStatus(id string, status booking.BookingStatus) error {
	query := `UPDATE bookings SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.Exec(context.Background(), query, string(status), id)
	return err
}

func (r *BookingRepository) GetByID(id string) (*booking.Booking, error) {
	query := `SELECT id, user_id, show_id, price, status, created_at, updated_at FROM bookings WHERE id = $1`
	var b booking.Booking
	var status string
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&b.ID,
		&b.UserID,
		&b.ShowID,
		&b.Price,
		&status,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	b.Status = booking.BookingStatus(status)

	seatQuery := `SELECT seat_id FROM booking_seats WHERE booking_id = $1`
	rows, err := r.db.Query(context.Background(), seatQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seatIDs []string
	for rows.Next() {
		var seatID string
		if err := rows.Scan(&seatID); err != nil {
			return nil, err
		}
		seatIDs = append(seatIDs, seatID)
	}
	b.SeatIDs = seatIDs

	return &b, nil
}

func (r *BookingRepository) GetByUserID(userID string) ([]*booking.Booking, error) {
	query := `SELECT id, user_id, show_id, price, status, created_at, updated_at FROM bookings WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(context.Background(), query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*booking.Booking
	for rows.Next() {
		var b booking.Booking
		var status string
		if err := rows.Scan(
			&b.ID,
			&b.UserID,
			&b.ShowID,
			&b.Price,
			&status,
			&b.CreatedAt,
			&b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		b.Status = booking.BookingStatus(status)
		bookings = append(bookings, &b)
	}

	for _, b := range bookings {
		seatQuery := `SELECT seat_id FROM booking_seats WHERE booking_id = $1`
		sRows, err := r.db.Query(context.Background(), seatQuery, b.ID)
		if err != nil {
			return nil, err
		}
		
		var seatIDs []string
		for sRows.Next() {
			var seatID string
			if err := sRows.Scan(&seatID); err != nil {
				sRows.Close()
				return nil, err
			}
			seatIDs = append(seatIDs, seatID)
		}
		sRows.Close()
		b.SeatIDs = seatIDs
	}

	return bookings, nil
}
