## errors
### 为什么使用errors
+ 基于[github.com/pkg/errors](https://github.com/pkg/errors)扩展，基本无上手成本和迁移成本。
+ 和qlog打通，开启accessLog的服务，自动打印error和error.stack。
+ 不会无限打印堆栈信息，有最大长度限制(5)。

### 导入包
```bash
import "github.com/leon-yc/ggs/pkg/errors"
```

### 如何使用
#### 创建
```bash
errors.New("message")
errors.Newf("format", args...)
errors.Errorf("format", args...)
```
#### 只追加堆栈信息
```bash
errors.WithStack(err)
```
#### 只追加上下文
```bash
errors.WithMessage(err, "message")
errors.WithMessagef(err, "format", args...)
```
#### 同时追加堆栈和上下文
```bash
errors.Wrap(err, "message")
errors.Wrapf(err, "format", args...)
```
#### 获取堆栈信息[和qlog打通，不需要手动调用]
```bash
errors.Stack()
```
#### Cause递归检索最原始的error
```bash
switch err := errors.Cause(err).(type) {
case *MyError:
        // handle specifically
default:
        // unknown error
}
```
#### 打印堆栈信息
```bash
qlog.WithError(err).Error("xxx")
qlog.WithField("error", err).Error("xxx")
qlog.WithFields(qlog.Fields{"error": err}).Error("xxx")
```
堆栈信息会显示在err.stack中，格式为file:line，对多显示5个。需要查看完整的调用链路请查看trace平台。