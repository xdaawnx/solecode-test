package cache

import (
	mocks "solecode/pkg/cache/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestUser struct tanpa JSON tags untuk matching yang lebih mudah
type TestUser struct {
	Name  string
	Email string
}

func TestWithGeneratedMock(t *testing.T) {
	mockCache := &mocks.CacheItf{}

	testKey := "test_key"
	testValue := "test_value"
	testUser := TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	t.Run("Mock Get and Set", func(t *testing.T) {
		// Setup expectations
		mockCache.On("Set", testKey, testValue, time.Minute).Return(nil)
		mockCache.On("Get", testKey).Return(testValue, nil)

		// Test Set
		err := mockCache.Set(testKey, testValue, time.Minute)
		assert.NoError(t, err)

		// Test Get
		result, err := mockCache.Get(testKey)
		assert.NoError(t, err)
		assert.Equal(t, testValue, result)

		// Verify expectations
		mockCache.AssertExpectations(t)
	})

	t.Run("Mock GetJSON and SetJSON dengan type yang tepat", func(t *testing.T) {
		// Setup expectations dengan type yang eksplisit
		mockCache.On("SetJSON", testKey, testUser, time.Minute).Return(nil)
		mockCache.On("GetJSON", testKey, mock.AnythingOfType("*cache.TestUser")).
			Run(func(args mock.Arguments) {
				// Set the value of the passed pointer
				v := args.Get(1).(*TestUser)
				v.Name = testUser.Name
				v.Email = testUser.Email
			}).
			Return(nil)

		// Test SetJSON
		err := mockCache.SetJSON(testKey, testUser, time.Minute)
		assert.NoError(t, err)

		// Test GetJSON
		var result TestUser
		err = mockCache.GetJSON(testKey, &result)
		assert.NoError(t, err)
		assert.Equal(t, testUser.Name, result.Name)
		assert.Equal(t, testUser.Email, result.Email)

		// Verify expectations
		mockCache.AssertExpectations(t)
	})

	t.Run("Mock GetJSON dan SetJSON dengan interface{}", func(t *testing.T) {
		testData := map[string]interface{}{
			"name":  "Jane Doe",
			"email": "jane@example.com",
		}

		// Setup expectations untuk interface{}
		mockCache.On("SetJSON", "user_data", testData, time.Minute).Return(nil)
		mockCache.On("GetJSON", "user_data", mock.AnythingOfType("*map[string]interface {}")).
			Run(func(args mock.Arguments) {
				v := args.Get(1).(*map[string]interface{})
				*v = testData
			}).
			Return(nil)

		// Test SetJSON
		err := mockCache.SetJSON("user_data", testData, time.Minute)
		assert.NoError(t, err)

		// Test GetJSON
		var result map[string]interface{}
		err = mockCache.GetJSON("user_data", &result)
		assert.NoError(t, err)
		assert.Equal(t, testData, result)

		// Verify expectations
		mockCache.AssertExpectations(t)
	})

	t.Run("Mock Delete", func(t *testing.T) {
		// Setup expectations
		mockCache.On("Delete", testKey).Return(nil)

		// Test Delete
		err := mockCache.Delete(testKey)
		assert.NoError(t, err)

		// Verify expectations
		mockCache.AssertExpectations(t)
	})

	t.Run("Mock Get returns nil for non-existent key", func(t *testing.T) {
		// Setup expectations
		mockCache.On("Get", "non_existent").Return(nil, nil)

		// Test Get for non-existent key
		result, err := mockCache.Get("non_existent")
		assert.NoError(t, err)
		assert.Nil(t, result)

		// Verify expectations
		mockCache.AssertExpectations(t)
	})

	t.Run("Mock error scenarios", func(t *testing.T) {
		expectedError := ErrCacheUnavailable

		// Setup expectations for error
		mockCache.On("Set", "error_key", "value", time.Minute).Return(expectedError)
		mockCache.On("Get", "error_key").Return(nil, expectedError)

		// Test Set with error
		err := mockCache.Set("error_key", "value", time.Minute)
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)

		// Test Get with error
		result, err := mockCache.Get("error_key")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedError, err)

		// Verify expectations
		mockCache.AssertExpectations(t)
	})
}

func TestGeneratedMockInUserService(t *testing.T) {
	mockCache := &mocks.CacheItf{}

	t.Run("User service cache operations", func(t *testing.T) {
		userData := map[string]interface{}{
			"id":    "user_123",
			"name":  "John Doe",
			"email": "john@example.com",
		}

		// Setup expectations for user service flow
		mockCache.On("SetJSON", "user:user_123", userData, time.Hour).Return(nil)
		mockCache.On("GetJSON", "user:user_123", mock.AnythingOfType("*map[string]interface {}")).
			Run(func(args mock.Arguments) {
				v := args.Get(1).(*map[string]interface{})
				*v = userData
			}).
			Return(nil)
		mockCache.On("Delete", "user:user_123").Return(nil)

		// Simulate user service operations
		err := mockCache.SetJSON("user:user_123", userData, time.Hour)
		assert.NoError(t, err)

		var retrievedUser map[string]interface{}
		err = mockCache.GetJSON("user:user_123", &retrievedUser)
		assert.NoError(t, err)
		assert.Equal(t, userData, retrievedUser)

		err = mockCache.Delete("user:user_123")
		assert.NoError(t, err)

		// Verify all expectations were met
		mockCache.AssertExpectations(t)
	})
}

func TestGeneratedMockWithArgumentMatchers(t *testing.T) {
	mockCache := &mocks.CacheItf{}

	t.Run("Using argument matchers", func(t *testing.T) {
		// Setup expectations with argument matchers
		mockCache.On("Set", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).
			Return(nil)
		mockCache.On("Get", mock.MatchedBy(func(key string) bool {
			return len(key) > 0
		})).Return("matched_value", nil)

		// Test with various arguments
		err := mockCache.Set("any_key", "any_value", time.Minute)
		assert.NoError(t, err)

		result, err := mockCache.Get("valid_key")
		assert.NoError(t, err)
		assert.Equal(t, "matched_value", result)

		mockCache.AssertExpectations(t)
	})

	t.Run("Using mock.Anything for flexible matching", func(t *testing.T) {
		// Setup dengan mock.Anything untuk lebih fleksibel
		mockCache.On("SetJSON", mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("time.Duration")).
			Return(nil)
		mockCache.On("GetJSON", mock.AnythingOfType("string"), mock.Anything).
			Return(nil)

		// Test dengan berbagai data
		err := mockCache.SetJSON("key1", "string_value", time.Minute)
		assert.NoError(t, err)

		err = mockCache.SetJSON("key2", 123, time.Minute)
		assert.NoError(t, err)

		err = mockCache.SetJSON("key3", map[string]interface{}{"field": "value"}, time.Minute)
		assert.NoError(t, err)

		var result interface{}
		err = mockCache.GetJSON("any_key", &result)
		assert.NoError(t, err)

		mockCache.AssertExpectations(t)
	})
}

func TestGeneratedMockWithComplexScenarios(t *testing.T) {
	mockCache := &mocks.CacheItf{}

	t.Run("Complex cache operations sequence", func(t *testing.T) {
		user1 := TestUser{Name: "User 1", Email: "user1@example.com"}
		user2 := TestUser{Name: "User 2", Email: "user2@example.com"}

		// Setup sequence of operations
		mockCache.On("SetJSON", "user:1", user1, time.Hour).Return(nil).Once()
		mockCache.On("SetJSON", "user:2", user2, time.Hour).Return(nil).Once()
		mockCache.On("GetJSON", "user:1", mock.AnythingOfType("*cache.TestUser")).
			Run(func(args mock.Arguments) {
				v := args.Get(1).(*TestUser)
				*v = user1
			}).
			Return(nil).Once()
		mockCache.On("GetJSON", "user:2", mock.AnythingOfType("*cache.TestUser")).
			Run(func(args mock.Arguments) {
				v := args.Get(1).(*TestUser)
				*v = user2
			}).
			Return(nil).Once()
		mockCache.On("Delete", "user:1").Return(nil).Once()
		mockCache.On("Delete", "user:2").Return(nil).Once()

		// Execute sequence
		err := mockCache.SetJSON("user:1", user1, time.Hour)
		assert.NoError(t, err)

		err = mockCache.SetJSON("user:2", user2, time.Hour)
		assert.NoError(t, err)

		var result1 TestUser
		err = mockCache.GetJSON("user:1", &result1)
		assert.NoError(t, err)
		assert.Equal(t, user1, result1)

		var result2 TestUser
		err = mockCache.GetJSON("user:2", &result2)
		assert.NoError(t, err)
		assert.Equal(t, user2, result2)

		err = mockCache.Delete("user:1")
		assert.NoError(t, err)

		err = mockCache.Delete("user:2")
		assert.NoError(t, err)

		// Verify all expectations
		mockCache.AssertExpectations(t)
	})
}
