-- +goose Up
-- +goose StatementBegin
CREATE TABLE storage (
    ID INT,
    Rate INT,
    Capacity INT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- +goose StatementEnd