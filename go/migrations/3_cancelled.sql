-- +migrate Up
CREATE TABLE `cancelled_class` (
	`id`			int UNSIGNED NOT NULL AUTO_INCREMENT,
	`cancelled` 	int NOT NULL,
	`place` 		int NOT NULL,
	`week` 			int NOT NULL,
	`period`		int NOT NULL,
	`day` 			varchar(255) NOt NULL,
	`class_name` 	varchar(255) NOT NULL,
	`instructor` 	varchar(255) NOT NULL,
	`reason_id` 	int NOT NULL,
	PRIMARY KEY(`id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

-- +migrate Down
DROP TABLE cancelled_class;
