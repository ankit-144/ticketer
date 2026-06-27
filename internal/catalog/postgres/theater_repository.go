package postgres

import (
	"context"
	"ticketer/internal/catalog"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TheaterRepository struct {
	db *pgxpool.Pool
}

func NewTheaterRepository(db *pgxpool.Pool) catalog.TheaterRepository {
	return &TheaterRepository{
		db: db,
	}
}


func (r *TheaterRepository) GetByID(id string) (*catalog.Theater, error) {
	query := `SELECT id, admin_id, name, location FROM theaters WHERE id = $1`
	var theater catalog.Theater
	err := r.db.QueryRow(context.Background(), query, id).Scan(
		&theater.ID,
		&theater.AdminID,
		&theater.Name,
		&theater.Location,
	)
	if err != nil {
		return nil, err
	}
	return &theater, nil
}

func (r *TheaterRepository) GetByAdminID(adminID string) ([]catalog.Theater, error) {
	query := `SELECT id, admin_id, name, location FROM theaters WHERE admin_id = $1`
	rows, err := r.db.Query(context.Background(), query, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var theaters []catalog.Theater
	for rows.Next() {
		var theater catalog.Theater
		if err := rows.Scan(
			&theater.ID,
			&theater.AdminID,
			&theater.Name,
			&theater.Location,
		); err != nil {
			return nil, err
		}
		
		theaters = append(theaters, theater)
	}
	return theaters, nil
}

func (r *TheaterRepository) GetScreen(screenID string) (*catalog.Screen, error) {
	query := `SELECT id, theater_id, name FROM screens WHERE id = $1`
	var screen catalog.Screen
	err := r.db.QueryRow(context.Background(), query, screenID).Scan(
		&screen.ID,
		&screen.TheaterID,
		&screen.Name,
	)
	if err != nil {
		return nil, err
	}

	return &screen, nil
}

func (r *TheaterRepository) GetScreens(theaterID string) ([]catalog.Screen, error) {
	query := `SELECT id, theater_id, name FROM screens WHERE theater_id = $1`
	rows, err := r.db.Query(context.Background(), query, theaterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var screens []catalog.Screen
	for rows.Next() {
		var screen catalog.Screen
		if err := rows.Scan(&screen.ID, &screen.TheaterID, &screen.Name); err != nil {
			return nil, err
		}
		screens = append(screens, screen)
	}
	return screens, nil
}

func (r *TheaterRepository) GetSeats(screenID string) ([]catalog.Seat, error) {
	query := `SELECT id, screen_id, row, number, type FROM seats WHERE screen_id = $1`
	rows, err := r.db.Query(context.Background(), query, screenID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []catalog.Seat
	for rows.Next() {
		var seat catalog.Seat
		var seatType string
		if err := rows.Scan(&seat.ID, &seat.ScreenID, &seat.Row, &seat.Number, &seatType); err != nil {
			return nil, err
		}
		seat.Type = catalog.SeatType(seatType)
		seats = append(seats, seat)
	}
	return seats, nil
}

func (r *TheaterRepository) GetSeatsByIDs(seatIDs []string) ([]catalog.Seat, error) {
	query := `SELECT id, screen_id, row, number, type FROM seats WHERE id = ANY($1)`
	rows, err := r.db.Query(context.Background(), query, seatIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seats []catalog.Seat
	for rows.Next() {
		var seat catalog.Seat
		var seatType string
		if err := rows.Scan(&seat.ID, &seat.ScreenID, &seat.Row, &seat.Number, &seatType); err != nil {
			return nil, err
		}
		seat.Type = catalog.SeatType(seatType)
		seats = append(seats, seat)
	}
	return seats, nil
}

func (r *TheaterRepository) List() ([]catalog.Theater, error) {
	query := `SELECT id, admin_id, name, location FROM theaters`
	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var theaters []catalog.Theater
	for rows.Next() {
		var theater catalog.Theater
		if err := rows.Scan(
			&theater.ID,
			&theater.AdminID,
			&theater.Name,
			&theater.Location,
		); err != nil {
			return nil, err
		}
		
		theaters = append(theaters, theater)
	}
	return theaters, nil
}

func (r *TheaterRepository) Save(theater *catalog.Theater) error {
	query := `INSERT INTO theaters (id, admin_id, name, location) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(context.Background(), query,
		theater.ID,
		theater.AdminID,
		theater.Name,
		theater.Location,
	)
	return err
}

func (r *TheaterRepository) Update(theater *catalog.Theater) error {
	query := `UPDATE theaters SET admin_id = $1, name = $2, location = $3 WHERE id = $4`
	_, err := r.db.Exec(context.Background(), query,
		theater.AdminID,
		theater.Name,
		theater.Location,
		theater.ID,
	)
	return err
}

func (r *TheaterRepository) Delete(id string) error {
	query := `DELETE FROM theaters WHERE id = $1`
	_, err := r.db.Exec(context.Background(), query, id)
	return err
}

func (r *TheaterRepository) AddScreenToTheater(theaterID string, screen *catalog.Screen) error {
	query := `INSERT INTO screens (id, theater_id, name) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(context.Background(), query,
		screen.ID,
		theaterID,
		screen.Name,
	)
	return err
}

func (r *TheaterRepository) UpdateScreen(screen *catalog.Screen) error {
	query := `UPDATE screens SET name = $1 WHERE id = $2`
	_, err := r.db.Exec(context.Background(), query,
		screen.Name,
		screen.ID,
	)
	return err
}

func (r *TheaterRepository) DeleteScreen(theaterID, screenID string) error {
	query := `DELETE FROM screens WHERE id = $1 AND theater_id = $2`
	_, err := r.db.Exec(context.Background(), query, screenID, theaterID)
	return err
}

func (r *TheaterRepository) AddSeatToScreen(screenID string, seat *catalog.Seat) ( error) {
	query := `INSERT INTO seats (id, screen_id, row, number, type) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(context.Background(), query,
		seat.ID,
		screenID,
		seat.Row,
		seat.Number,
		string(seat.Type),
	)

	return err
}

func (r *TheaterRepository) UpdateSeat(seat *catalog.Seat) error {
	query := `UPDATE seats SET row = $1, number = $2, type = $3 WHERE id = $4`
	_, err := r.db.Exec(context.Background(), query,
		seat.Row,
		seat.Number,
		string(seat.Type),
		seat.ID,
	)
	return err
}

func (r *TheaterRepository) DeleteSeat(screenID, seatID string) error {
	query := `DELETE FROM seats WHERE id = $1 AND screen_id = $2`
	_, err := r.db.Exec(context.Background(), query, seatID, screenID)
	return err
}
