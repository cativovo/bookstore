-- +goose Up
-- +goose StatementBegin
ALTER TABLE genre ADD CONSTRAINT unique_name UNIQUE (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE genre DROP CONSTRAINT unique_name;
-- +goose StatementEnd
