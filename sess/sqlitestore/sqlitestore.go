package sqlitestore

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

// SQLiteStore represents the session store.
type SQLiteStore struct {
	readDB      *sql.DB
	writeDB     *sql.DB
	stopCleanup chan bool
}

// NewSQLiteStore returns a new SQLiteStore instance, with a background cleanup goroutine
// that runs every 5 minutes to remove expired session data.
func NewSQLiteStore(readDB *sql.DB, writeDB *sql.DB) *SQLiteStore {
	return NewSQLiteStoreWithCleanupInterval(readDB, writeDB, 5*time.Minute)
}

// NewSQLiteStoreWithCleanupInterval returns a new SQLiteStore instance. The cleanupInterval
// parameter controls how frequently expired session data is removed by the
// background cleanup goroutine. Setting it to 0 prevents the cleanup goroutine
// from running (i.e. expired sessions will not be removed).
func NewSQLiteStoreWithCleanupInterval(readDB *sql.DB, writeDB *sql.DB, cleanupInterval time.Duration) *SQLiteStore {
	p := &SQLiteStore{readDB: readDB, writeDB: writeDB}
	if cleanupInterval > 0 {
		go p.startCleanup(cleanupInterval)
	}
	return p
}

// Find returns the data for a given session token from the SQLiteStore instance.
// If the session token is not found or is expired, the returned exists flag will
// be set to false.
func (p *SQLiteStore) Find(token string) (b []byte, exists bool, err error) {
	row := p.readDB.QueryRow("SELECT data FROM sessions WHERE token = $1 AND JULIANDAY('now') < expiry", token)
	err = row.Scan(&b)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	}
	return b, true, nil
}

// Commit adds a session token and data to the SQLiteStore instance with the
// given expiry time. If the session token already exists, then the data and expiry
// time are updated.
func (p *SQLiteStore) Commit(token string, b []byte, expiry time.Time) error {
	_, err := p.writeDB.Exec("REPLACE INTO sessions (token, data, expiry) VALUES ($1, $2, JULIANDAY($3))", token, b, expiry.UTC().Format("2006-01-02T15:04:05.999"))
	if err != nil {
		return err
	}
	return nil
}

// Delete removes a session token and corresponding data from the SQLiteStore
// instance.
func (p *SQLiteStore) Delete(token string) error {
	_, err := p.writeDB.Exec("DELETE FROM sessions WHERE token = $1", token)
	return err
}

// All returns a map containing the token and data for all active (i.e.
// not expired) sessions in the SQLiteStore instance.
func (p *SQLiteStore) All() (map[string][]byte, error) {
	rows, err := p.readDB.Query("SELECT token, data FROM sessions WHERE JULIANDAY('now') < expiry")
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	sessions := make(map[string][]byte)

	for rows.Next() {
		var (
			token string
			data  []byte
		)

		err = rows.Scan(&token, &data)
		if err != nil {
			return nil, err
		}

		sessions[token] = data
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return sessions, nil
}

func (p *SQLiteStore) startCleanup(interval time.Duration) {
	p.stopCleanup = make(chan bool)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			err := p.deleteExpired()
			if err != nil {
				log.Println(err)
			}
		case <-p.stopCleanup:
			ticker.Stop()
			return
		}
	}
}

// StopCleanup terminates the background cleanup goroutine for the SQLiteStore
// instance. It's rare to terminate this; generally SQLiteStore instances and
// their cleanup goroutines are intended to be long-lived and run for the lifetime
// of your application.
//
// There may be occasions though when your use of the SQLiteStore is transient.
// An example is creating a new SQLiteStore instance in a test function. In this
// scenario, the cleanup goroutine (which will run forever) will prevent the
// SQLiteStore object from being garbage collected even after the test function
// has finished. You can prevent this by manually calling StopCleanup.
func (p *SQLiteStore) StopCleanup() {
	if p.stopCleanup != nil {
		p.stopCleanup <- true
	}
}

func (p *SQLiteStore) deleteExpired() error {
	_, err := p.writeDB.Exec("DELETE FROM sessions WHERE expiry < JULIANDAY('now')")
	return err
}
