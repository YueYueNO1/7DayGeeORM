package session

import "geeorm/log"

// 用于启动一个数据库事务
func (s *Session) Begin() (err error) {
	log.Info("transaction begin")
	//判断是否成功启动一个数据库事务
	if s.tx, err = s.db.Begin(); err != nil {
		log.Error(err)
		return
	}
	return
}

// 判断事务是否提交成功
func (s *Session) Commit() (err error) {
	log.Info("transaction commit")
	if err = s.tx.Commit(); err != nil {
		log.Error(err)
	}
	return
}

// 判断事务是否回滚成功
func (s *Session) Rollback() (err error) {

	log.Info("transaction rollback")
	if err = s.tx.Rollback(); err != nil {
		log.Error(err)
	}
	return

}
