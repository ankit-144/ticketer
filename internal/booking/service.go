package booking

type BookingService struct {
	// Add dependencies here (e.g., repositories, pricing service)
}

func NewBookingService() *BookingService {
	return &BookingService{}
}

type Service interface {
    
    InitiateBooking(userID string, showID string, seatIDs []string) (*Booking, error)
    
    ConfirmBooking(bookingID string) error
    
    CancelBooking(bookingID string) error
}



