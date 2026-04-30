DROP TABLE IF EXISTS `message`;

CREATE TABLE `message` (
  `ID` int(11) NOT NULL AUTO_INCREMENT,
  `sender_id` varchar(100) NOT NULL,
  `timestamp` timestamp NOT NULL,
  `content` varchar(10000) NOT NULL,
  PRIMARY KEY (`ID`)
) ENGINE=InnoDB AUTO_INCREMENT=53646 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_uca1400_ai_ci;