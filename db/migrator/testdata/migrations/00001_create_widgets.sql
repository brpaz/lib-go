-- +goose Up
-- +goose StatementBegin
CREATE TABLE widgets (id INTEGER PRIMARY KEY);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE widgets;
-- +goose StatementEnd
