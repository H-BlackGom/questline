package service

import "github.com/H-BlackGom/questline/internal/repository"

func Bootstrap(dbPath string) (*Services, error) {
	repo, err := repository.New(dbPath)
	if err != nil {
		return nil, err
	}

	questRepo := repository.NewQuestRepository(repo)
	playerRepo := repository.NewPlayerRepository(repo)

	playerService := NewPlayerService(playerRepo, repo)
	questService := NewQuestService(questRepo, playerRepo, repo)
	syncService := NewSyncService(repo)

	services := &Services{
		repo:   repo,
		Quest:  questService,
		Sync:   syncService,
		Player: playerService,
	}

	if _, err := services.Sync.EvaluateLazySync(); err != nil {
		_ = services.Close()
		return nil, err
	}

	return services, nil
}
