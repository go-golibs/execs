package execs

// ExitEvent - событие завершения процесса Runner
type ExitEvent struct {
	Member Member
	Err    error
}

// ExitTrace трассировка статусов завершений процессов
type ExitTrace []ExitEvent
