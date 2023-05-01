package outbox

import "errors"

var ErrNoRecrods = errors.New("no records in the outbox")
