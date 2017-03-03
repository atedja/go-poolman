package poolman

import (
	"errors"
)

var ErrInvalidWorkerCountOrQueueSize = errors.New("Worker count or queue size must be at least 1.")
