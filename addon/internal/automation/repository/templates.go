package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/micro-ha/mikrotik-presence/addon/internal/automation/domain"
)

func (r *Repository) ListTemplates(
	ctx context.Context,
	search string,
	category string,
) ([]domain.CapabilityTemplate, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, data FROM capability_templates ORDER BY id`) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	search = strings.ToLower(strings.TrimSpace(search))
	category = strings.ToLower(strings.TrimSpace(category))
	items := make([]domain.CapabilityTemplate, 0)
	for rows.Next() {
		var (
			id      string
			encoded string
		)
		if err := rows.Scan(&id, &encoded); err != nil {
			return nil, err
		}
		item, err := decodeTemplate(id, encoded)
		if err != nil {
			if r.logger != nil {
				r.logger.Warn("failed to decode capability template", "id", id, "err", err)
			}
			continue
		}
		if !matchesTemplateFilter(item, search, category) {
			continue
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) GetTemplate(ctx context.Context, id string) (domain.CapabilityTemplate, error) {
	var encoded string
	err := r.db.QueryRowContext(
		ctx,
		`SELECT data FROM capability_templates WHERE id = ?`,
		id,
	).Scan(&encoded)
	if errorsIsNotFound(err) {
		return domain.CapabilityTemplate{}, ErrNotFound
	}
	if err != nil {
		return domain.CapabilityTemplate{}, err
	}
	return decodeTemplate(id, encoded)
}

func (r *Repository) CreateTemplate(ctx context.Context, template domain.CapabilityTemplate) error {
	encoded, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("encode template: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = r.db.ExecContext(
		ctx,
		`INSERT INTO capability_templates(id, data, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		template.ID,
		string(encoded),
		now,
		now,
	)
	return err
}

func (r *Repository) UpdateTemplate(ctx context.Context, template domain.CapabilityTemplate) error {
	encoded, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("encode template: %w", err)
	}
	res, err := r.db.ExecContext(
		ctx,
		`UPDATE capability_templates SET data = ?, updated_at = ? WHERE id = ?`,
		string(encoded),
		time.Now().UTC().Format(time.RFC3339Nano),
		template.ID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteTemplate(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM capability_templates WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func decodeTemplate(id string, encoded string) (domain.CapabilityTemplate, error) {
	var template domain.CapabilityTemplate
	if err := json.Unmarshal([]byte(encoded), &template); err != nil {
		return domain.CapabilityTemplate{}, err
	}
	if strings.TrimSpace(template.ID) == "" {
		template.ID = id
	}
	return template, nil
}

func matchesTemplateFilter(item domain.CapabilityTemplate, search string, category string) bool {
	if category != "" && strings.ToLower(strings.TrimSpace(item.Category)) != category {
		return false
	}
	if search == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{item.ID, item.Label, item.Description}, " "))
	return strings.Contains(haystack, search)
}

func errorsIsNotFound(err error) bool {
	return err == sql.ErrNoRows
}
