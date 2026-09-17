package poslocal

import "testing"

func TestSaveOperationIsIdempotent(t *testing.T) {
	db, err := OpenStore(t.TempDir() + "/pos.db")
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	if err := SaveOperation(db, "op-1", "1001", "10", "PYG", "PENDING", "2026-09-16"); err != nil {
		t.Fatal(err)
	}
	if err := SaveOperation(db, "op-1", "1001", "10", "PYG", "PENDING", "2026-09-16"); err != nil {
		t.Fatal(err)
	}
	if err := SaveOperation(db, "op-1", "1001", "11", "PYG", "PENDING", "2026-09-16"); err != ErrOperationConflict {
		t.Fatal("changed operation was accepted")
	}
}

func TestSaveOperationRejectsNumberReuse(t *testing.T) {
	db, err := OpenStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	if err := SaveOperation(db, "op-1", "1001", "10", "PYG", "PENDING", "2026-09-16"); err != nil {
		t.Fatal(err)
	}
	if err := SaveOperation(db, "op-2", "1001", "10", "PYG", "PENDING", "2026-09-16"); err != ErrOperationConflict {
		t.Fatalf("number reuse accepted: %v", err)
	}
}

func TestSaveOperationRejectsInvalidLocalRecord(t *testing.T) {
	db, err := OpenStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	if err := SaveOperation(db, "", "1", "10", "PYG", "PENDING", "2026-09-16"); err != ErrInvalidOperation {
		t.Fatalf("expected invalid operation, got %v", err)
	}
	if err := SaveOperation(db, "op-1", "1", "10", "PYG", "UNKNOWN", "2026-09-16"); err != ErrInvalidOperation {
		t.Fatalf("expected invalid state, got %v", err)
	}
}

func TestSQLiteSchemaRejectsUnknownState(t *testing.T) {
	db, err := OpenStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer CloseAgentDB(db)
	err = db.Exec(`INSERT INTO pos_operations(operation_id,number,total,currency,state,created_at) VALUES('raw-1','1','10','PYG','UNKNOWN','2026-09-16')`).Error
	if err == nil {
		t.Fatal("schema accepted unknown operation state")
	}
}
