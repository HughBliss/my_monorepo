package tablizer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"github.com/hughbliss/my_protobuf/go/pkg/gen/tools/fields_translations_map"
	"github.com/hughbliss/my_toolkit/reporter"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
	"slices"
	"strings"
	"time"
)

var ALLOWED_SEPARATORS = []string{
	",", ";", "comma", "semicolon",
}

func New() *Service {
	return &Service{
		rep:     reporter.InitReporter("TablizerService"),
		storage: NewTempStorage(),
	}
}

type Service struct {
	rep     reporter.Reporter
	storage *TempStorage
}

func (s *Service) EchoMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx, log, end := s.rep.Start(c.Request().Context(), "EchoMiddleware")
			defer end()

			if v := c.Request().Header.Get("X-Tablizer-Enable"); v == "" {
				return next(c)
			}

			spanCTX := trace.SpanContextFromContext(ctx)
			if !spanCTX.HasTraceID() {
				return next(c)
			}

			capture := NewResponseCapture(c.Response().Writer)
			c.Response().Writer = capture

			ctx = context.WithValue(ctx, "tablizerEnable", true)

			if v := c.Request().Header.Get("X-Tablizer-Separator"); v != "" {
				ctx = context.WithValue(ctx, "tablizerSeparator", v)
			}

			c.SetRequest(c.Request().WithContext(ctx))
			if err := next(c); err != nil {
				return err
			}

			table, has := s.storage.Get(spanCTX.TraceID().String())
			if !has {
				log.Warn().Msg("There is not table in storage")
				capture.FlushOriginal()
				return nil
			}

			capture.FlushCSV(table)
			return nil

		}
	}
}

func (s *Service) GRPCMiddleware() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, _, end := s.rep.Start(ctx, "GRPCMiddleware")
		defer end()

		if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
			return err
		}

		if enabled, ok := ctx.Value("tablizerEnable").(bool); !ok || !enabled {
			return nil
		}

		separator := ","
		if sep, ok := ctx.Value("tablizerSeparator").(string); ok && slices.Contains(ALLOWED_SEPARATORS, sep) {
			separator = sep
		}

		spanCTX := trace.SpanContextFromContext(ctx)
		if !spanCTX.HasTraceID() {
			return nil
		}

		message, ok := reply.(protoreflect.ProtoMessage)
		if !ok {
			return nil
		}

		var methodTranslations, doTranslate = fields_translations_map.Map[method]
		var table [][]string
		message.ProtoReflect().Range(
			func(fieldDescriptor protoreflect.FieldDescriptor, value protoreflect.Value) bool {
				if fieldDescriptor == nil {
					return true
				}
				if !fieldDescriptor.IsList() {
					return true
				}
				if fieldDescriptor.Kind() != protoreflect.MessageKind {
					return true
				}

				list := value.List()
				if list.Len() == 0 || !list.IsValid() {
					return true
				}

				rowDescriptor := fieldDescriptor.Message()
				fields := rowDescriptor.Fields()

				var orderedFieldNames []string
				var header []string

				var tableTranslations map[string]string
				if doTranslate {
					tableTranslations, doTranslate = methodTranslations[string(fieldDescriptor.Name())]
				}
				for i := 0; i < fields.Len(); i++ {
					field := fields.Get(i)
					fieldName := string(field.Name())
					orderedFieldNames = append(orderedFieldNames, fieldName)
					if doTranslate {
						if translatedField, ok := tableTranslations[fieldName]; ok {
							fieldName = translatedField
						}
					}
					header = append(header, fieldName)
				}
				table = append(table, header)

				for i := 0; i < list.Len(); i++ {
					rowMsg := list.Get(i).Message()
					var row []string
					for _, fieldName := range orderedFieldNames {
						fieldDesc := rowDescriptor.Fields().ByName(protoreflect.Name(fieldName))
						value := rowMsg.Get(fieldDesc)
						row = append(row, formatValue(value))
					}
					table = append(table, row)
				}
				return false
			},
		)
		var buffer = new(bytes.Buffer)
		writer := csv.NewWriter(buffer)
		writer.Comma = separatorRune(separator)
		if err := writer.WriteAll(table); err != nil {
			return status.Error(codes.Internal, "при записи csv таблицы возникла ошибка")
		}

		s.storage.Store(spanCTX.TraceID().String(), buffer.Bytes())

		return nil
	}
}
func formatValue(v protoreflect.Value) string {
	i := v.Interface()
	switch val := i.(type) {
	case float64, float32:
		return strings.Replace(fmt.Sprintf("%.2f", val), ".", ",", 1)
	case bool:
		if val {
			return "да"
		}
		return "нет"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)
	case string:
		return val

	case []byte:
		return base64.StdEncoding.EncodeToString(val)
	case protoreflect.EnumNumber:

		return fmt.Sprintf("%d", v.Enum()) // TODO
	case protoreflect.List:
		var list []string
		for i := 0; i < val.Len(); i++ {
			list = append(list, formatValue(val.Get(i)))
		}
		return strings.Join(list, ", ")

	}
	if ts, ok := v.Message().Interface().(*timestamppb.Timestamp); ok {
		if err := ts.CheckValid(); err == nil {
			moscowTime := ts.AsTime().In(time.FixedZone("MSK", 3*60*60))
			return moscowTime.Format("15:04:05 02.01.2006 (МСК)")
		}
	}
	return ""

}
func separatorRune(s string) rune {
	switch s {
	case ",":
		return ','
	case "comma":
		return ','
	case ";":
		return ';'
	case "semicolon":
		return ';'
	}
	return ','
}
