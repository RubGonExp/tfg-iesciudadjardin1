package main

import "fmt"

// Storage es una envoltura para operaciones combinadas de cache y base de datos
type Storage struct {
	sqlstorage SQLStorage
	cache      *Cache
}

// Init inicia el conector de la base de datos
func (s *Storage) Init(user, password, host, name, redisHost, redisPort string, cache bool) error {
	if err := s.sqlstorage.Init(user, password, host, name); err != nil {
		return err
	}

	var err error
	s.cache, err = NewCache(redisHost, redisPort, cache)
	if err != nil {
		return err
	}

	return nil
}

// List recupera una lista de todos de la cache, si esta almacenada, o de la base de datos
func (s Storage) List() (Todos, error) {
	ts, err := s.cache.List()
	if err != nil {
		if err == ErrCacheMiss {
			ts, err = s.sqlstorage.List()
			if err != nil {
				return ts, fmt.Errorf("error getting list of todos from database: %v", err)
			}
		}
		if err := s.cache.SaveList(ts); err != nil {
			return ts, fmt.Errorf("error caching list of todos : %v", err)
		}
	}

	return ts, nil
}

// Create registra una nueva tarea en la base de datos.
func (s Storage) Create(t Todo) (Todo, error) {
	if err := s.cache.DeleteList(); err != nil {
		return Todo{}, fmt.Errorf("error clearing cache : %v", err)
	}

	t, err := s.sqlstorage.Create(t)
	if err != nil {
		return t, err
	}

	if err = s.cache.Save(t); err != nil {
		return t, err
	}

	return t, nil
}

// Read devuelve una sola tarea de la cache o de la base de datos
func (s Storage) Read(id string) (Todo, error) {
	t, err := s.cache.Get(id)
	if err != nil {
		if err == ErrCacheMiss {
			t, err = s.sqlstorage.Read(id)
			if err != nil {
				return t, fmt.Errorf("error getting single from database todo: %v", err)
			}
		}
		if err := s.cache.Save(t); err != nil {
			return t, fmt.Errorf("error caching single todo : %v", err)
		}
	}

	return t, nil
}

// Update modifica una tarea en la base de datos.
func (s Storage) Update(t Todo) error {
	if err := s.cache.DeleteList(); err != nil {
		return fmt.Errorf("error clearing cache : %v", err)
	}

	if err := s.sqlstorage.Update(t); err != nil {
		return err
	}

	if err := s.cache.Save(t); err != nil {
		return err
	}

	return nil
}

// Delete elimina una tarea de la base de datos.
func (s Storage) Delete(id string) error {
	if err := s.cache.DeleteList(); err != nil {
		return fmt.Errorf("error clearing cache : %v", err)
	}

	if err := s.sqlstorage.Delete(id); err != nil {
		return err
	}

	if err := s.cache.Delete(id); err != nil {
		return err
	}

	return nil
}
