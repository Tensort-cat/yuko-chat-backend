-- 联系申请表
CREATE TABLE contact_applies (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增id',
    uuid CHAR(20) NOT NULL COMMENT '申请id',
    user_id CHAR(20) NOT NULL COMMENT '申请人id',
    contact_id CHAR(20) NOT NULL COMMENT '被申请id',
    contact_type TINYINT NOT NULL COMMENT '被申请类型，0.用户，1.群聊',
    status TINYINT NOT NULL COMMENT '申请状态，0.申请中，1.通过，2.拒绝，3.拉黑',
    message VARCHAR(100) DEFAULT NULL COMMENT '申请信息',
    last_apply_at DATETIME NOT NULL COMMENT '最后申请时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    UNIQUE KEY idx_uuid (uuid),
    KEY idx_user_id (user_id),
    KEY idx_contact_id (contact_id),
    KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='联系申请表';

-- 群组信息表
CREATE TABLE group_infos (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增id',
    uuid CHAR(20) NOT NULL COMMENT '群组唯一id',
    name VARCHAR(20) NOT NULL COMMENT '群名称',
    notice VARCHAR(500) DEFAULT NULL COMMENT '群公告',
    members JSON DEFAULT NULL COMMENT '群组成员',
    member_cnt INT DEFAULT 1 COMMENT '群人数',
    owner_id CHAR(20) NOT NULL COMMENT '群主uuid',
    add_mode TINYINT DEFAULT 0 COMMENT '加群方式，0.直接，1.审核',
    avatar CHAR(255) NOT NULL DEFAULT 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png' COMMENT '头像',
    status TINYINT DEFAULT 0 COMMENT '状态，0.正常，1.禁用，2.解散',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    UNIQUE KEY idx_uuid (uuid),
    KEY idx_created_at (created_at),
    KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='群组信息表';

-- 消息表
CREATE TABLE messages (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '自增id',
  `uuid` CHAR(20) NOT NULL COMMENT '消息uuid',
  `session_id` CHAR(20) NOT NULL COMMENT '会话uuid',
  `type` TINYINT NOT NULL COMMENT '消息类型，0.文本，1.语音，2.文件，3.通话',
  `content` TEXT COMMENT '消息内容',
  `url` CHAR(255) COMMENT '消息url',
  `send_id` CHAR(20) NOT NULL COMMENT '发送者uuid',
  `send_name` VARCHAR(20) NOT NULL COMMENT '发送者昵称',
  `send_avatar` VARCHAR(255) NOT NULL COMMENT '发送者头像',
  `receive_id` CHAR(20) NOT NULL COMMENT '接受者uuid',
  `file_type` CHAR(10) COMMENT '文件类型',
  `file_name` VARCHAR(50) COMMENT '文件名',
  `file_size` CHAR(20) COMMENT '文件大小',
  `status` TINYINT NOT NULL COMMENT '状态，0.未发送，1.已发送',
  `created_at` DATETIME NOT NULL COMMENT '创建时间',
  `send_at` DATETIME NULL COMMENT '发送时间',
  `av_data` TEXT COMMENT '通话传递数据',

  PRIMARY KEY (`id`),

  UNIQUE KEY `uk_uuid` (`uuid`),

  KEY `idx_session_id` (`session_id`),
  KEY `idx_send_id` (`send_id`),
  KEY `idx_receive_id` (`receive_id`)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息表';

-- 会话表
CREATE TABLE sessions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增id',
    uuid CHAR(20) NOT NULL COMMENT '会话uuid',
    send_id CHAR(20) NOT NULL COMMENT '创建会话人id',
    receive_id CHAR(20) NOT NULL COMMENT '接受会话人id',
    receive_name VARCHAR(20) NOT NULL COMMENT '名称',
    avatar CHAR(255) NOT NULL DEFAULT 'default_avatar.png' COMMENT '头像',
    created_at DATETIME DEFAULT NULL COMMENT '创建时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    UNIQUE KEY idx_uuid (uuid),
    KEY idx_send_id (send_id),
    KEY idx_receive_id (receive_id),
    KEY idx_created_at (created_at),
    KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='会话表';

-- 用户联系表
CREATE TABLE user_contacts (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增id',
    user_id CHAR(20) NOT NULL COMMENT '用户唯一id',
    contact_id CHAR(20) NOT NULL COMMENT '对应联系id',
    contact_type TINYINT NOT NULL COMMENT '联系类型，0.用户，1.群聊',
    status TINYINT NOT NULL COMMENT '联系状态，0.正常，1.拉黑，2.被拉黑，3.删除好友，4.被删除好友，5.被禁言，6.退出群聊，7.被踢出群聊',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    update_at DATETIME NOT NULL COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    KEY idx_user_id (user_id),
    KEY idx_contact_id (contact_id),
    KEY idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户联系表';

-- 用户信息表
CREATE TABLE user_infos (
    id BIGINT AUTO_INCREMENT PRIMARY KEY COMMENT '自增id',
    uuid CHAR(20) NOT NULL COMMENT '用户唯一id',
    nickname VARCHAR(20) NOT NULL COMMENT '昵称',
    telephone CHAR(11) NOT NULL COMMENT '电话',
    email CHAR(30) DEFAULT NULL COMMENT '邮箱',
    avatar CHAR(255) NOT NULL DEFAULT 'https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png' COMMENT '头像',
    gender TINYINT DEFAULT NULL COMMENT '性别，0.男，1.女',
    signature VARCHAR(100) DEFAULT NULL COMMENT '个性签名',
    password LONGTEXT NOT NULL COMMENT '密码',
    birthday CHAR(8) DEFAULT NULL COMMENT '生日',
    created_at DATETIME NOT NULL COMMENT '创建时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '删除时间',
    is_admin TINYINT NOT NULL COMMENT '是否是管理员，0.不是，1.是',
    status TINYINT NOT NULL COMMENT '状态，0.正常，1.禁用',
    UNIQUE KEY idx_uuid (uuid),
    KEY idx_telephone (telephone),
    KEY idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户信息表';