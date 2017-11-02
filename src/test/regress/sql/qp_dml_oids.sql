create schema qp_dml_oids;
set search_path='qp_dml_oids';

DROP TABLE IF EXISTS dml_ao;
CREATE TABLE dml_ao (a int , b int default -1, c text) WITH (appendonly = true, oids = true) DISTRIBUTED BY (a);
INSERT INTO dml_ao VALUES(generate_series(1,2),generate_series(1,2),'r');

INSERT INTO dml_ao VALUES(NULL,NULL,NULL);

--
-- Check that a tuple gets an OID, even if it's toasted (there used to
-- be a bug, where toasting a tuple cleared its just-assigned OID)
--
INSERT INTO dml_ao (a, b, c) VALUES (10, 1, repeat('x', 50000));
INSERT INTO dml_ao (a, b, c) VALUES (10, 2, repeat('x', 50000));

SELECT COUNT(distinct oid) FROM dml_ao where a = 10;