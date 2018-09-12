CREATE TABLE `user` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `userID` varchar(32) NOT NULL,
  `watch` varchar(45) DEFAULT NULL,
  `post` varchar(45) DEFAULT NULL,
  `feed` varchar(45) DEFAULT NULL,
  `createDate` date DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `userID_UNIQUE` (`userID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;


CREATE TABLE `post` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `text` varchar(256) DEFAULT NULL,
  `date` datetime DEFAULT NULL,
  `stars` int(11) DEFAULT NULL,
  `comments` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8;

