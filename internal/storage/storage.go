package storage

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"order/internal/entity"
)

// Реализация репозитория
type Storage struct {
	db *sql.DB
}

func NewStorage(db *sql.DB) Store {
	return &Storage{db: db}
}

func (s *Storage) SaveOrder(ctx context.Context, order entity.Order) error {

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("Failed to rollback: %v", err)
		}
	}()

	// Вставка в таблицу orders
	_, err = tx.ExecContext(ctx, `
        INSERT INTO orders (
            order_uid, track_number, entry, locale, internal_signature,
            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		log.Printf("Failed to insert order: %v", err)
		return err
	}

	// Вставка в таблицу deliveries
	_, err = tx.ExecContext(ctx, `
        INSERT INTO deliveries (
            order_uid, name, phone, zip, city, address, region, email
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		log.Printf("Failed to insert delivery: %v", err)
		return err
	}

	// Вставка в таблицу payments
	_, err = tx.ExecContext(ctx, `
        INSERT INTO payments (
            transaction, order_uid, request_id, currency, provider, amount,
            payment_dt, bank, delivery_cost, goods_total, custom_fee
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (transaction) DO NOTHING`,
		order.Payment.Transaction, order.OrderUID, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		log.Printf("Failed to insert payment: %v", err)
		return err
	}

	// Вставка в таблицу items
	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, `
            INSERT INTO items (
                order_uid, chrt_id, track_number, price, rid, name,
                sale, size, total_price, nm_id, brand, status
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
            ON CONFLICT DO NOTHING`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			log.Printf("Failed to insert item: %v", err)
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		return err
	}

	log.Printf("Order %s saved successfully", order.OrderUID)
	return nil
}

func (s *Storage) GetOrder(ctx context.Context, orderUID string) (entity.Order, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT 
            o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
            o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
            d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
            p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
            p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
            i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, i.size,
            i.total_price, i.nm_id, i.brand, i.status
        FROM orders o
        INNER JOIN deliveries d ON o.order_uid = d.order_uid
        INNER JOIN payments p ON o.order_uid = p.order_uid
        INNER JOIN items i ON o.order_uid = i.order_uid
        WHERE o.order_uid = $1
        ORDER BY o.order_uid`, orderUID)
	if err != nil {
		return entity.Order{}, fmt.Errorf("failed to query order %s: %v", orderUID, err)
	}
	defer rows.Close()

	var order entity.Order
	var items []entity.Item
	found := false

	for rows.Next() {
		var item entity.Item
		err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
			&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
			&order.Delivery.Email,
			&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
			&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
			&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
			&order.Payment.CustomFee,
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return entity.Order{}, fmt.Errorf("failed to scan order %s: %v", orderUID, err)
		}

		found = true
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return entity.Order{}, fmt.Errorf("error iterating rows for order %s: %v", orderUID, err)
	}

	if !found {
		return entity.Order{}, fmt.Errorf("order %s not found", orderUID)
	}

	order.Items = items
	return order, nil
}

func (s *Storage) GetAllOrders(ctx context.Context) ([]entity.Order, error) {
	rows, err := s.db.QueryContext(ctx, `
        SELECT 
            o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
            o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
            d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
            p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
            p.bank, p.delivery_cost, p.goods_total, p.custom_fee,
            i.chrt_id, i.track_number, i.price, i.rid, i.name, i.sale, i.size,
            i.total_price, i.nm_id, i.brand, i.status
        FROM orders o
        INNER JOIN deliveries d ON o.order_uid = d.order_uid
        INNER JOIN payments p ON o.order_uid = p.order_uid
        INNER JOIN items i ON o.order_uid = i.order_uid
        ORDER BY o.order_uid`)
	if err != nil {
		return nil, fmt.Errorf("failed to query all orders: %v", err)
	}
	defer rows.Close()

	ordersMap := make(map[string]*entity.Order)
	for rows.Next() {
		var order entity.Order
		var item entity.Item
		err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale,
			&order.InternalSignature, &order.CustomerID, &order.DeliveryService,
			&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
			&order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip,
			&order.Delivery.City, &order.Delivery.Address, &order.Delivery.Region,
			&order.Delivery.Email,
			&order.Payment.Transaction, &order.Payment.RequestID, &order.Payment.Currency,
			&order.Payment.Provider, &order.Payment.Amount, &order.Payment.PaymentDt,
			&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal,
			&order.Payment.CustomFee,
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name,
			&item.Sale, &item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %v", err)
		}

		existingOrder, exists := ordersMap[order.OrderUID]
		if !exists {
			order.Items = []entity.Item{item}
			ordersMap[order.OrderUID] = &order
		} else {
			existingOrder.Items = append(existingOrder.Items, item)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	var orders []entity.Order
	for _, order := range ordersMap {
		orders = append(orders, *order)
	}

	if len(orders) == 0 {
		log.Printf("No complete orders found in database")
	}

	return orders, nil
}
