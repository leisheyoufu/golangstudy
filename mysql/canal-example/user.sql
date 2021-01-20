create database if not exists Test;
use Test;
create table Test.User
(
  id int auto_increment primary key,
  name varchar(40) null,
  status enum("active","deleted") DEFAULT "active",
  created timestamp default CURRENT_TIMESTAMP not null on update CURRENT_TIMESTAMP
) engine=InnoDB;

INSERT Into Test.User (`id`,`name`) VALUE (1,"Jack");
UPDATE Test.User SET name="Jonh" WHERE id=1;
DELETE FROM Test.User WHERE id=1;


create table NewUser(user1 varchar(40), user2 varchar(40),user3 varchar(40), user4 varchar(40), user5 varchar(40), user6 varchar(40), user7 varchar(40), user8 varchar(40));
insert into NewUser VALUES("1", "2","3","4","5","6","7","8"),("1", "2","3","4","5","6","7","8");

SELECT version() AS TITLE;
select column_name,column_key from information_schema.columns where table_schema='Test' and table_name='NewUser'


CREATE TABLE `t1` (
  `ID` int(4),
  `name` varchar(40),
  UNIQUE KEY `catename` (`name`)  
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
insert into t1 value(1, "1");