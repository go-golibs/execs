package execs

// Member реализация участника группы раннеров
type Member struct {
	Name string
	Runner
}

// Members группа раннеров
type Members []Member
