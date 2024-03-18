CREATE TABLE trading.positions (
    operation_id uuid PRIMARY KEY,
    user_id uuid,
    symbol varchar(10),
    open_price decimal,
    close_price decimal,
    buy boolean,
    FOREIGN KEY (user_id) REFERENCES trading.balance(id) ON DELETE CASCADE
)