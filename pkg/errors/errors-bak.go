package errors

//
//import (
//	"fmt"
//	"google.golang.org/grpc/status"
//	"net/http"
//	"runtime"
//)
//
//type Coder interface {
//	Code() int
//	HTTPStatus() int
//	String() string
//	Reference() string
//}
//
//var coders = make(map[int]Coder)
//
//func MustRegister(coder Coder) {
//	if coder == nil {
//		return
//	}
//	code := coder.Code()
//	if _, ok := coders[code]; ok {
//		panic(fmt.Sprintf("error code %d already registered", code))
//	}
//	coders[code] = coder
//}
//
//func register(code int, httpStatus int, message string, refs ...string) {
//	var ref string
//	if len(refs) > 0 {
//		ref = refs[0]
//	}
//	coder := ErrCode{
//		C:    code,
//		HTTP: httpStatus,
//		Ext:  message,
//		Ref:  ref,
//	}
//	MustRegister(coder)
//}
//
//type ErrCode struct {
//	C    int
//	HTTP int
//	Ext  string
//	Ref  string
//}
//
//func (e ErrCode) Code() int {
//	if e.C == 0 {
//		return http.StatusInternalServerError
//	}
//	return e.C
//}
//
//func (e ErrCode) HTTPStatus() int {
//	return e.HTTP
//}
//
//func (e ErrCode) String() string {
//	return e.Ext
//}
//
//func (e ErrCode) Reference() string {
//	return e.Ref
//}
//
//var _ Coder = (*ErrCode)(nil)
//
//func ParseCoder(err error) Coder {
//	if err == nil {
//		return nil
//	}
//	if coder, ok := coders[errorCode(err)]; ok {
//		return coder
//	}
//	return nil
//}
//
//func errorCode(err error) int {
//	if err == nil {
//		return 0
//	}
//	type coder interface {
//		Code() int
//	}
//	if e, ok := err.(coder); ok {
//		return e.Code()
//	}
//	if e, ok := err.(*withCode); ok {
//		return e.code
//	}
//	return 0
//}
//
//type withCode struct {
//	code    int
//	message string
//	cause   error
//	*stack
//}
//
//func (e *withCode) Error() string {
//	if e.cause != nil {
//		return fmt.Sprintf("%s: %v", e.message, e.cause)
//	}
//	return e.message
//}
//
//func (e *withCode) Code() int {
//	return e.code
//}
//
//func (e *withCode) Unwrap() error {
//	return e.cause
//}
//
//type withMessage struct {
//	cause error
//	msg   string
//}
//
//func (e *withMessage) Error() string {
//	return e.msg
//}
//
//func (e *withMessage) Cause() error {
//	return e.cause
//}
//
//func (e *withMessage) Unwrap() error {
//	return e.cause
//}
//
//type fundamental struct {
//	msg string
//	*stack
//}
//
//func (f *fundamental) Error() string { return f.msg }
//
//type stack []uintptr
//
//func callers() *stack {
//	const depth = 32
//	var pcs [depth]uintptr
//	n := runtime.Callers(3, pcs[:])
//	st := stack(pcs[:n])
//	return &st
//}
//
//func New(message string) error {
//	return &fundamental{
//		msg:   message,
//		stack: callers(),
//	}
//}
//
//func Errorf(format string, args ...interface{}) error {
//	return &fundamental{
//		msg:   fmt.Sprintf(format, args...),
//		stack: callers(),
//	}
//}
//
//func WithCode(code int, format string, args ...interface{}) error {
//	return &withCode{
//		code:    code,
//		message: fmt.Sprintf(format, args...),
//		stack:   callers(),
//	}
//}
//
//func Wrap(err error, message string) error {
//	if err == nil {
//		return nil
//	}
//	if e, ok := err.(*withCode); ok {
//		return &withCode{
//			code:    e.code,
//			message: message,
//			cause:   err,
//			stack:   callers(),
//		}
//	}
//	return &withMessage{
//		cause: err,
//		msg:   message,
//	}
//}
//
//func Wrapf(err error, format string, args ...interface{}) error {
//	if err == nil {
//		return nil
//	}
//	if e, ok := err.(*withCode); ok {
//		return &withCode{
//			code:    e.code,
//			message: fmt.Sprintf(format, args...),
//			cause:   err,
//			stack:   callers(),
//		}
//	}
//	return &withMessage{
//		cause: err,
//		msg:   fmt.Sprintf(format, args...),
//	}
//}
//
//func WithMessage(err error, message string) error {
//	if err == nil {
//		return nil
//	}
//	return &withMessage{
//		cause: err,
//		msg:   message,
//	}
//}
//
//func WithStack(err error) error {
//	if err == nil {
//		return nil
//	}
//	return &withMessage{
//		cause: err,
//		msg:   err.Error(),
//	}
//}
//
//func Cause(err error) error {
//	type causer interface {
//		Cause() error
//	}
//	for err != nil {
//		cause, ok := err.(causer)
//		if !ok {
//			break
//		}
//		err = cause.Cause()
//	}
//	return err
//}
//
//func IsCode(err error, code int) bool {
//	return errorCode(err) == code
//}
//
//func FromGrpcError(e error) error {
//	if e == nil {
//		return e
//	}
//
//	st, ok := status.FromError(e)
//	if !ok {
//		return WithCode(100002, "unknown error")
//	}
//
//	return &withCode{
//		err:  st.Err(),
//		code: int(st.Code()),
//	}
//}
