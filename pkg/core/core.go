package core

type RouterRegister interface {
	Register(c *GHPRContext) error
}
