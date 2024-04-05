CREATE TABLE trading.positions (
    operation_id uuid PRIMARY KEY,
    user_id uuid,
    symbol varchar(10),
    open_price decimal,
    close_price decimal,
    created_at timestamp WITH time zone DEFAULT NOW(),
    long boolean
);

CREATE FUNCTION capture_func() 
RETURNS trigger AS 
$$
DECLARE
  payload json;
BEGIN
    IF (TG_OP = 'INSERT') THEN
		payload = format('{"symbol":"%s","user_id":"%s","open_price":"%s","long":"%s"}', NEW.symbol, NEW.user_id, NEW.open_price, NEW.long);
        RAISE NOTICE '%', 'NOTIFY on INSERT';
        EXECUTE FORMAT('NOTIFY positionOpen, ''%s''', payload);
        RETURN NEW;
    ELSE
		payload = format('{"symbol":"%s","user_id":"%s","close_price":"%s"}', NEW.symbol, NEW.user_id, NEW.close_price);
        RAISE NOTICE '%', 'NOTIFY on UPDATE';
        EXECUTE FORMAT('NOTIFY positionClose, ''%s''', payload);
		RETURN NEW;
    END IF;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER mytrigger BEFORE INSERT OR UPDATE
    ON  trading.positions FOR EACH ROW EXECUTE PROCEDURE capture_func();