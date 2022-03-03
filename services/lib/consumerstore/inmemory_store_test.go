package consumerstore

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInMemoryConsumerApiProvider_LogConsume(t *testing.T) {
	ctx := context.Background()

	store := NewInMemoryConsumerApiProvider()
	err := store.LogConsume(ctx, &AppLogConsumeProgress{})
	assert.EqualError(t, err, "missing consumer id")

	err = store.LogConsume(ctx, &AppLogConsumeProgress{ConsumerId: "one"})
	assert.EqualError(t, err, "missing offset")

	err = store.LogConsume(ctx, &AppLogConsumeProgress{ConsumerId: "one", Offset: "22"})
	assert.NoError(t, err)

	progress, err := store.GetLogConsume(ctx, "one")
	assert.NoError(t, err)
	assert.NotNil(t, progress)
	assert.Equal(t, progress.ConsumerId, "one")
	assert.Equal(t, progress.Offset, "22")

	err = store.LogConsume(ctx, &AppLogConsumeProgress{ConsumerId: "one", Offset: "33"})
	assert.NoError(t, err)

	progress, err = store.GetLogConsume(ctx, "one")
	assert.NoError(t, err)
	assert.NotNil(t, progress)
	assert.Equal(t, progress.ConsumerId, "one")
	assert.Equal(t, progress.Offset, "33")

	t.Run("add another consumer", func(tt *testing.T) {
		err = store.LogConsume(ctx, &AppLogConsumeProgress{ConsumerId: "two", Offset: "10"})
		assert.NoError(t, err)

		progress, err = store.GetLogConsume(ctx, "two")
		assert.NoError(t, err)
		assert.NotNil(t, progress)
		assert.Equal(t, progress.ConsumerId, "two")
		assert.Equal(t, progress.Offset, "10")
	})

	t.Run("listing", func(tt *testing.T) {
		consumers, err := store.List(ctx)
		assert.NoError(tt, err)
		assert.Len(tt, consumers, 2)

		expected := map[string]string{
			"one": "33",
			"two": "10",
		}

		found := map[string]struct{}{}
		for _, c := range consumers {
			assert.Equal(tt, expected[c.ConsumerId], c.Offset)
			found[c.ConsumerId] = struct{}{}
		}

		assert.Len(tt, found, len(expected), "some entries were not found")
	})
}
