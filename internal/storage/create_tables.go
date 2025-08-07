package storage

import (
	"context"
	"database/sql"
	"log"
)

// InitDB создаёт таблицы и индексы отдельными запросами в транзакции
func InitDB(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("Failed to rollback: %v", err)
		}
	}()

	// Создание таблицы orders
	_, err = tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS orders (
            order_uid TEXT PRIMARY KEY,
            track_number TEXT,
            entry TEXT,
            locale TEXT,
            internal_signature TEXT,
            customer_id TEXT,
            delivery_service TEXT,
            shardkey TEXT,
            sm_id INTEGER,
            date_created TIMESTAMP,
            oof_shard TEXT
        )`)
	if err != nil {
		log.Printf("Failed to create orders table: %v", err)
		return err
	}

	// Создание таблицы deliveries
	_, err = tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS deliveries (
            id SERIAL PRIMARY KEY,
            order_uid TEXT UNIQUE,
            name TEXT,
            phone TEXT,
            zip TEXT,
            city TEXT,
            address TEXT,
            region TEXT,
            email TEXT,
            FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
        )`)
	if err != nil {
		log.Printf("Failed to create deliveries table: %v", err)
		return err
	}

	// Создание таблицы payments
	_, err = tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS payments (
            transaction TEXT PRIMARY KEY,
            order_uid TEXT UNIQUE,
            request_id TEXT,
            currency TEXT,
            provider TEXT,
            amount INTEGER,
            payment_dt BIGINT,
            bank TEXT,
            delivery_cost INTEGER,
            goods_total INTEGER,
            custom_fee INTEGER,
            FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
        )`)
	if err != nil {
		log.Printf("Failed to create payments table: %v", err)
		return err
	}

	// Создание таблицы items
	_, err = tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS items (
            id SERIAL PRIMARY KEY,
            order_uid TEXT,
            chrt_id BIGINT,
            track_number TEXT,
            price INTEGER,
            rid TEXT,
            name TEXT,
            sale INTEGER,
            size TEXT,
            total_price INTEGER,
            nm_id INTEGER,
            brand TEXT,
            status INTEGER,
            FOREIGN KEY (order_uid) REFERENCES orders(order_uid) ON DELETE CASCADE
        )`)
	if err != nil {
		log.Printf("Failed to create items table: %v", err)
		return err
	}

	// Создание индексов
	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_delivery_order_uid ON deliveries(order_uid)`)
	if err != nil {
		log.Printf("Failed to create index idx_delivery_order_uid: %v", err)
		return err
	}

	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_payment_order_uid ON payments(order_uid)`)
	if err != nil {
		log.Printf("Failed to create index idx_payment_order_uid: %v", err)
		return err
	}

	_, err = tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items(order_uid)`)
	if err != nil {
		log.Printf("Failed to create index idx_items_order_uid: %v", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}
