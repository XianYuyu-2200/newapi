package model

import (
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedeemBalanceCodeAddsQuota(t *testing.T) {
	truncateTables(t)

	insertUserForPaymentGuardTest(t, 501, 100)
	code := &Redemption{
		UserId:      1,
		Key:         "balance-redemption-key",
		Status:      common.RedemptionCodeStatusEnabled,
		Name:        "Balance Code",
		Type:        RedemptionTypeBalance,
		Quota:       250,
		CreatedTime: common.GetTimestamp(),
	}
	require.NoError(t, code.Insert())

	result, err := Redeem(code.Key, 501)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, RedemptionTypeBalance, result.Type)
	assert.Equal(t, 250, result.Quota)
	assert.Zero(t, result.PlanId)
	assert.Zero(t, result.SubscriptionId)
	assert.Equal(t, 350, getUserQuotaForPaymentGuardTest(t, 501))
	assert.Zero(t, countUserSubscriptionsForPaymentGuardTest(t, 501))
}

func TestRedeemSubscriptionCodeCreatesActiveSubscriptionWithoutAddingQuota(t *testing.T) {
	truncateTables(t)

	insertUserForPaymentGuardTest(t, 502, 100)
	plan := insertSubscriptionPlanForPaymentGuardTest(t, 601)
	code := &Redemption{
		UserId:      1,
		Key:         "subscription-redemption-key",
		Status:      common.RedemptionCodeStatusEnabled,
		Name:        "Subscription Code",
		Type:        RedemptionTypeSubscription,
		PlanId:      plan.Id,
		Quota:       999,
		CreatedTime: common.GetTimestamp(),
	}
	require.NoError(t, code.Insert())

	result, err := Redeem(code.Key, 502)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, RedemptionTypeSubscription, result.Type)
	assert.Zero(t, result.Quota)
	assert.Equal(t, plan.Id, result.PlanId)
	assert.NotZero(t, result.SubscriptionId)
	assert.Equal(t, 100, getUserQuotaForPaymentGuardTest(t, 502))

	var subscription UserSubscription
	require.NoError(t, DB.Where("id = ?", result.SubscriptionId).First(&subscription).Error)
	assert.Equal(t, 502, subscription.UserId)
	assert.Equal(t, plan.Id, subscription.PlanId)
	assert.Equal(t, "active", subscription.Status)
	assert.Equal(t, "redemption", subscription.Source)

	var redeemed Redemption
	require.NoError(t, DB.Where("id = ?", code.Id).First(&redeemed).Error)
	assert.Equal(t, common.RedemptionCodeStatusUsed, redeemed.Status)
	assert.Equal(t, 502, redeemed.UsedUserId)
	assert.NotZero(t, redeemed.RedeemedTime)
}
