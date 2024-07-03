package post

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type User struct {
	Id             int
	PassportSerie  int
	PassportNumber int
	Surname        string
	Name           string
	Patronymic     string
	Address        string
}

func (pg *postgres) CreateUser(ctx context.Context, passportSerie, passportNumber int, surname, name, patronymic string, address string) (int, error) {
	query := `
	INSERT INTO users (passport_serie, passport_number, surname, name, patronymic, address) 
	VALUES (@passport_serie, @passport_number, @surname, @name, @patronymic, @address) RETURNING id`

	args := pgx.NamedArgs{
		"passport_serie":  passportSerie,
		"passport_number": passportNumber,
		"surname":         surname,
		"name":            name,
		"patronymic":      patronymic,
		"address":         address,
	}

	result := pg.db.QueryRow(ctx, query, args)

	var id int
	err := result.Scan(&id)

	if err != nil {
		return -1, fmt.Errorf("unable to insert row: %w", err)
	}

	return id, nil
}

func (pg *postgres) GetUser(ctx context.Context, id *int, passportSerie, passportNumber *int, surname, name, patronymic *string, address *string, offset, limit *int) ([]User, error) {
	query := `
	select *
	from users `

	if id != nil || passportNumber != nil || passportSerie != nil || surname != nil || name != nil || patronymic != nil || address != nil{
		query += ` where `
	}

	query, args := makeQueryArgs(query, id, passportNumber, passportSerie, surname, name, patronymic, address, offset, limit)

	rows, err := pg.db.Query(ctx, query, args)

	if err != nil {
		return nil, err
	}

	defer rows.Close()
	result, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}

	return result, err
}

func (pg *postgres) DeleteUser(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = @id`

	args := pgx.NamedArgs{
		"id": id,
	}

	results, err := pg.db.Exec(ctx, query, args)

	if results.RowsAffected() == 0 {
		return errors.New("not id delet")
	}

	if err != nil {
		return fmt.Errorf("unable to delete content: %w", err)
	}

	return nil
}

func (pg *postgres) UpdateUser(ctx context.Context, id int, passportSerie, passportNumber int, surname, name, patronymic string, address string) error {
	query := `
	UPDATE users SET passportSerie = @passportSerie, passportNumber = @passportNumber,
	surname = @surname, name = @name, patronymic = @patronymic, address = @address 
	WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": id,
		"passportSerie": passportSerie,
		"passportNumber": passportNumber,
		"surname": surname,
		"name": name,
		"patronymic": patronymic,
		"address": address,
	}

	results, err := pg.db.Exec(ctx, query, args)

	if results.RowsAffected() == 0 {
		return errors.New("not id delet")
	}

	if err != nil {
		return fmt.Errorf("unable to delete content: %w", err)
	}

	return nil
}


func makeQueryArgs(query string, id *int, passportNumber, passportSerie *int, surname, name, patronymic *string, address *string, offset, limit *int) (string, pgx.NamedArgs) {
	namedAM := make(map[string]any)
	count := 0

	if id != nil {
		query += withAnd(" id = @id ", count)
		namedAM["id"] = *id
		count++
	}
	if passportNumber != nil {
		query += withAnd(" passport_number = @passport_number ", count)
		namedAM["passport_number"] = *passportNumber
		count++
	}
	if passportSerie != nil {
		query += withAnd(" passport_serie = @passport_serie ", count)
		namedAM["passport_serie"] = *passportSerie
		count++
	}
	if surname != nil {
		query += withAnd(" surname = @surname  ", count)
		namedAM["surname"] = *surname
		count++
	}
	if name != nil {
		query += withAnd(" name = @name  ", count)
		namedAM["name"] = *name
		count++
	}
	if patronymic != nil {
		query += withAnd(" patronymic = @patronymic ", count)
		namedAM["patronymic"] = *patronymic
		count++
	}
	if address != nil {
		query += withAnd(" address = @address ", count)
		namedAM["address"] = *address
		count++
	}
	if offset != nil {
		query += " offset = @offset "
		namedAM["offset"] = *offset
	}
	if limit != nil {
		query += " limit = @limit "
		namedAM["limit"] = *limit
	}
	return query, pgx.NamedArgs(namedAM)
}


func withAnd(query string, count int) string{
	if count > 0 {
		return " AND " + query
	}
	return query
}
