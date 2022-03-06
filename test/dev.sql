select * from accounts;

---

select * from metadata;

---

update accounts set active = 0 WHERE id = 1;

---

SELECT * FROM accounts WHERE address_index IN(1,5, 7);

---

update metadata set value = '0' where key = 'last_height';

---

SELECT SUM(entries) FROM accounts WHERE active = 1

---

SELECT a.id, a.user_address, a.user_name, COUNT(e.id) as wins
FROM entries AS e
LEFT JOIN accounts as a ON a.id = e.account_id
WHERE e.id IN (3,4,5)
GROUP BY a.id

---

SELECT user_address, COUNT(id) as c, SUM(amount) FROM accounts GROUP BY 1 HAVING c > 1

---

UPDATE accounts SET
active = 0,
user_name = NULL,
user_address = NULL,
amount = 0,
entries = 0,
ref_id = 0
WHERE user_address IN (SELECT user_address FROM accounts GROUP BY 1 HAVING COUNT(id) > 1)
    AND amount = 0;

---

SELECT * FROM accounts;