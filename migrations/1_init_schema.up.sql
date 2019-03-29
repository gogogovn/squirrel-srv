CREATE TABLE countries
(
  id         INT(11)      NOT NULL PRIMARY KEY AUTO_INCREMENT,
  created_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at DATETIME              DEFAULT NULL,
  name       VARCHAR(255) NOT NULL,
  code       VARCHAR(255) NOT NULL,
  UNIQUE KEY uid_code (code),
  KEY idx_name (name)
);

CREATE TABLE vpn_servers
(
  id               INT(11)  NOT NULL PRIMARY KEY AUTO_INCREMENT,
  created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at       DATETIME          DEFAULT NULL,
  host_name        VARCHAR(255),
  ip               VARCHAR(255),
  score            BIGINT,
  ping             INT(11),
  speed            BIGINT,
  country_id       INT(11),
  num_vpn_sessions BIGINT,
  uptime           BIGINT,
  total_users      BIGINT,
  total_traffic    BIGINT,
  log_type         VARCHAR(255),
  operator         VARCHAR(255),
  message          VARCHAR(255),
  open_vpn_config  TEXT
);