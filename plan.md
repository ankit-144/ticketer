# Ticketer: System Design & Architecture

## Core Requirements
- View movies in theaters.
- Book tickets (Select movie, theater, show, seats).
- Display seating arrangements.
- Handle payments and confirm bookings.
- Manage concurrent bookings (Real-time seat availability).
- Support dynamic pricing (Seat types, show timings).
- Admin management for movies, shows, and theaters.

## Architecture: Clean Architecture / Layered Approach
The system follows a layered architecture to ensure separation of concerns and maintainability.

### 1. Domain Models (`internal/models`)
- **Entities:** `Movie`, `Theater`, `Screen`, `Seat`, `Show`, `Booking`, `User`.
- **Logic:** Minimal business logic, mostly data structures.

### 2. Repository Layer (Data Access)
Responsible for persistence (DB operations). Each entity has its own repository interface.

```go
type MovieRepository interface {
    GetByID(id string) (*Movie, error)
    List() ([]Movie, error)
    Save(movie *Movie) error
}

type TheaterRepository interface {
    GetByID(id string) (*Theater, error)
    GetScreen(screenID string) (*Screen, error)
    List() ([]Theater, error)
}

type ShowRepository interface {
    GetByID(id string) (*Show, error)
    GetByMovie(movieID string) ([]Show, error)
    GetByTheater(theaterID string) ([]Show, error)
}

type BookingRepository interface {
    Create(booking *Booking) error
    UpdateStatus(id string, status BookingStatus) error
    GetByID(id string) (*Booking, error)
}
```

### 3. Service Layer (Business Logic)
This is where the core logic resides. Following SRP, each service handles a specific domain.

#### `BookingService` (SRP: Orchestrates the booking process)
```go
type BookingService interface {
    // Initiate booking, reserve seats (temporary lock)
    InitiateBooking(userID string, showID string, seatIDs []string) (*Booking, error)
    // Finalize booking after payment
    ConfirmBooking(bookingID string) error
    // Handle expiration of temporary locks
    CancelBooking(bookingID string) error
}
```

#### `AvailabilityService` (SRP: Manages seat state and concurrency)
```go
type AvailabilityService interface {
    GetAvailableSeats(showID string) ([]Seat, error)
    LockSeats(showID string, seatIDs []string) error
    ReleaseSeats(showID string, seatIDs []string) error
}
```

#### `PricingService` (SRP: Calculates dynamic pricing)
```go
type PricingService interface {
    CalculatePrice(movie Movie, show Show, seats []Seat) (float64, error)
}
```

### 4. Controller/Handler Layer
Exposes functionality via REST/gRPC.

---

## Logical Relationships & Design Decisions

### 1. Entity Relationships
- **Movie 1:N Show:** A movie can have many shows across different theaters.
- **Theater 1:N Screen:** A theater consists of multiple screens.
- **Screen 1:N Seat:** A physical screen has a fixed layout of seats.
- **Show 1:1 Movie & 1:1 Screen:** A show is a specific movie on a specific screen at a specific time.
- **Booking 1:N Seat:** A single booking can include multiple seats for a specific show.

### 2. Concurrency Control (Crucial)
- **Problem:** Two users booking the same seat simultaneously.
- **Solution:** 
    - Use **Pessimistic Locking** (e.g., Redis-based distributed locks or DB `SELECT FOR UPDATE`) when a user selects a seat.
    - Seats should be in a "Locked/Reserved" state for X minutes until payment is confirmed or the lock expires.

### 3. Separation of Concerns (SRP)
- **Pricing is decoupled:** The `PricingService` handles the logic for calculating costs (Base Price + Seat Type Premium + Time Surcharge), so `Movie` or `Seat` models don't need to know *how* they are priced.
- **Availability is decoupled:** `AvailabilityService` manages the real-time state of seats during a show, separate from the permanent `Booking` records.

### 4. Dependency Inversion
- Services should depend on Interfaces (Repositories), not concrete implementations. This allows for easy swapping of databases (e.g., SQL to NoSQL) or mocking for unit tests.

## Development Workflow
1. Implement Repositories (In-memory first, then DB).
2. Implement `AvailabilityService` with locking logic.
3. Implement `PricingService`.
4. Implement `BookingService` to coordinate everything.
5. Add API Handlers.
