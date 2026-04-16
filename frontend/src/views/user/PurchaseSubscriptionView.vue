<template>
  <AppLayout>
    <div class="mx-auto max-w-6xl space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex justify-center py-12">
        <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary-500 border-t-transparent"></div>
      </div>

      <!-- Feature Disabled State -->
      <div v-else-if="!purchaseEnabled" class="card p-12 text-center">
        <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700">
          <Icon name="creditCard" size="xl" class="text-gray-400" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
          自助购买功能未开启
        </h3>
        <p class="text-gray-500 dark:text-dark-400">
          管理员尚未启用自助订阅购买功能，请联系管理员。
        </p>
      </div>

      <!-- Empty State -->
      <div v-else-if="plans.length === 0" class="card p-12 text-center">
        <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-gray-100 dark:bg-dark-700">
          <Icon name="creditCard" size="xl" class="text-gray-400" />
        </div>
        <h3 class="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
          暂无可购买套餐
        </h3>
        <p class="text-gray-500 dark:text-dark-400">
          当前没有可购买的订阅套餐，请稍后再来查看。
        </p>
      </div>

      <!-- Plan Cards Grid -->
      <template v-else>
        <div class="grid gap-6 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
          <div
            v-for="plan in plans"
            :key="plan.group_id"
            class="card overflow-hidden flex flex-col"
          >
            <!-- Card Header -->
            <div class="border-b border-gray-100 p-5 dark:border-dark-700">
              <div class="flex items-center justify-between">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                  {{ plan.display_name }}
                </h3>
                <span
                  v-if="plan.is_subscribed"
                  class="inline-flex items-center rounded-full bg-emerald-100 px-2.5 py-0.5 text-xs font-medium text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400"
                >
                  已订阅
                </span>
              </div>
              <p v-if="plan.description" class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                {{ plan.description }}
              </p>
            </div>

            <!-- Price Section -->
            <div class="px-5 pt-4">
              <div class="flex items-baseline gap-2">
                <template v-if="plan.discounted_price != null && plan.discounted_price < plan.price">
                  <span class="text-2xl font-bold text-primary-600 dark:text-primary-400">
                    ¥{{ plan.discounted_price.toFixed(2) }}
                  </span>
                  <span class="text-sm text-gray-400 line-through dark:text-dark-500">
                    ¥{{ plan.price.toFixed(2) }}
                  </span>
                  <span class="inline-flex items-center rounded-full bg-orange-100 px-2 py-0.5 text-xs font-medium text-orange-700 dark:bg-orange-900/30 dark:text-orange-400">
                    VIP 优惠
                  </span>
                </template>
                <template v-else>
                  <span class="text-2xl font-bold text-gray-900 dark:text-white">
                    ¥{{ plan.price.toFixed(2) }}
                  </span>
                </template>
              </div>
              <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">
                有效期 {{ plan.validity_days }} 天
              </p>
              <p v-if="plan.current_expiry" class="mt-0.5 text-xs text-emerald-600 dark:text-emerald-400">
                当前到期：{{ formatDateOnly(new Date(plan.current_expiry)) }}
              </p>
            </div>

            <!-- Limits & Details -->
            <div class="flex-1 space-y-3 px-5 py-4">
              <!-- Usage Limits -->
              <div class="space-y-1.5">
                <div v-if="plan.daily_limit_usd" class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                  <svg class="h-4 w-4 flex-shrink-0 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>
                  日限额 ${{ plan.daily_limit_usd.toFixed(2) }}
                </div>
                <div v-if="plan.weekly_limit_usd" class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                  <svg class="h-4 w-4 flex-shrink-0 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>
                  周限额 ${{ plan.weekly_limit_usd.toFixed(2) }}
                </div>
                <div v-if="plan.monthly_limit_usd" class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                  <svg class="h-4 w-4 flex-shrink-0 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>
                  月限额 ${{ plan.monthly_limit_usd.toFixed(2) }}
                </div>
                <div v-if="!plan.daily_limit_usd && !plan.weekly_limit_usd && !plan.monthly_limit_usd" class="flex items-center gap-2 text-sm text-emerald-600 dark:text-emerald-400">
                  <span class="text-lg">∞</span>
                  无用量限制
                </div>
              </div>

              <!-- Rate Multiplier -->
              <div v-if="plan.rate_multiplier !== 1" class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400">
                <svg class="h-4 w-4 flex-shrink-0 text-primary-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                费率倍率 {{ plan.rate_multiplier }}x
              </div>

              <!-- Supported Models -->
              <div v-if="plan.supported_model_scopes && plan.supported_model_scopes.length > 0" class="space-y-1">
                <p class="text-xs font-medium text-gray-500 dark:text-dark-400">支持模型</p>
                <div class="flex flex-wrap gap-1">
                  <span
                    v-for="scope in plan.supported_model_scopes.slice(0, 5)"
                    :key="scope"
                    class="inline-flex rounded-md bg-gray-100 px-2 py-0.5 text-xs text-gray-600 dark:bg-dark-700 dark:text-gray-400"
                  >
                    {{ scope }}
                  </span>
                  <span
                    v-if="plan.supported_model_scopes.length > 5"
                    class="inline-flex rounded-md bg-gray-100 px-2 py-0.5 text-xs text-gray-500 dark:bg-dark-700 dark:text-gray-400"
                  >
                    +{{ plan.supported_model_scopes.length - 5 }}
                  </span>
                </div>
              </div>

              <!-- Features List -->
              <div v-if="plan.features && plan.features.length > 0" class="space-y-1.5 pt-1">
                <div
                  v-for="(feature, idx) in plan.features"
                  :key="idx"
                  class="flex items-start gap-2 text-sm text-gray-600 dark:text-gray-400"
                >
                  <svg class="mt-0.5 h-4 w-4 flex-shrink-0 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                  </svg>
                  {{ feature }}
                </div>
              </div>
            </div>

            <!-- Action Button -->
            <div class="border-t border-gray-100 p-5 dark:border-dark-700">
              <button
                type="button"
                class="btn btn-primary w-full py-2.5"
                @click="openPurchaseDialog(plan)"
              >
                {{ plan.is_subscribed ? '续费' : '购买' }}
              </button>
            </div>
          </div>
        </div>
      </template>

      <!-- Payment Confirmation Dialog -->
      <teleport to="body">
        <transition name="modal">
          <div v-if="showPurchaseDialog" class="fixed inset-0 z-50 flex items-center justify-center">
            <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="closePurchaseDialog"></div>
            <div class="relative mx-4 w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl dark:bg-dark-800">
              <div class="flex items-start justify-between">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">确认购买</h3>
                <button type="button" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300" @click="closePurchaseDialog">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              <div class="mt-4 space-y-4">
                <!-- Plan Details -->
                <div class="rounded-xl bg-gray-50 p-4 dark:bg-dark-700">
                  <p class="font-medium text-gray-900 dark:text-white">{{ selectedPlan?.display_name }}</p>
                  <p class="mt-1 text-sm text-gray-500 dark:text-dark-400">有效期 {{ selectedPlan?.validity_days }} 天</p>
                  <div class="mt-3 space-y-1 text-sm">
                    <div class="flex justify-between">
                      <span class="text-gray-500 dark:text-dark-400">原价</span>
                      <span class="text-gray-700 dark:text-gray-300">¥{{ selectedPlan?.price.toFixed(2) }}</span>
                    </div>
                    <div v-if="discountAmount > 0" class="flex justify-between">
                      <span class="text-orange-600 dark:text-orange-400">VIP 优惠</span>
                      <span class="text-orange-600 dark:text-orange-400">-¥{{ discountAmount.toFixed(2) }}</span>
                    </div>
                    <div class="flex justify-between border-t border-gray-200 pt-1 dark:border-dark-600">
                      <span class="font-medium text-gray-900 dark:text-white">应付金额</span>
                      <span class="font-bold text-primary-600 dark:text-primary-400">¥{{ finalPrice.toFixed(2) }}</span>
                    </div>
                  </div>
                </div>

                <!-- Payment Method Selection -->
                <div>
                  <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">支付方式</label>
                  <div class="space-y-2">
                    <!-- Balance Payment -->
                    <button
                      type="button"
                      :class="[
                        'flex w-full items-center gap-3 rounded-xl border-2 px-4 py-3 transition-all text-left',
                        purchaseMethod === 'balance'
                          ? 'border-primary-500 bg-primary-50 dark:border-primary-400 dark:bg-primary-900/20'
                          : 'border-gray-200 bg-white hover:border-gray-300 dark:border-dark-600 dark:bg-dark-800 dark:hover:border-dark-500'
                      ]"
                      @click="purchaseMethod = 'balance'"
                    >
                      <Icon name="creditCard" size="md" class="flex-shrink-0 text-gray-400" />
                      <div class="flex-1">
                        <span class="text-sm font-medium text-gray-700 dark:text-gray-300">余额支付</span>
                        <span class="ml-2 text-xs text-gray-500 dark:text-dark-400">
                          (余额: ${{ user?.balance?.toFixed(2) || '0.00' }})
                        </span>
                      </div>
                    </button>

                    <!-- Online Payment Methods -->
                    <button
                      v-for="method in onlinePaymentMethods"
                      :key="method.value"
                      type="button"
                      :class="[
                        'flex w-full items-center gap-3 rounded-xl border-2 px-4 py-3 transition-all text-left',
                        purchaseMethod === method.value
                          ? 'border-primary-500 bg-primary-50 dark:border-primary-400 dark:bg-primary-900/20'
                          : 'border-gray-200 bg-white hover:border-gray-300 dark:border-dark-600 dark:bg-dark-800 dark:hover:border-dark-500'
                      ]"
                      @click="purchaseMethod = method.value"
                    >
                      <!-- Alipay icon -->
                      <svg v-if="isAlipayMethod(method.value)" class="h-5 w-5 flex-shrink-0" viewBox="0 0 24 24">
                        <circle cx="12" cy="12" r="10" fill="#1677FF"/>
                        <path fill="white" d="M18.36 16.83c-1.05-.46-3.2-1.38-5.27-2.39.9-1.24 1.62-2.66 2.1-4.2H12V8.86h3.6V7.85H12V5.47a.41.41 0 00-.41-.41h-1.33v2.79H7v1.01h3.22v1.38H7.08v1h6.88c-.38 1.2-.93 2.3-1.62 3.2-1.68-.71-3.5-1.31-5.1-1.31-2.48 0-4.08 1.18-4.08 3.05 0 1.96 1.54 3.38 4.59 3.38 2.24 0 4.29-.97 5.87-2.52 1.81.93 4.62 2.2 5.2 2.46.2.1.4 0 .51-.18l.5-1.05c.1-.17 0-.4-.21-.5z"/>
                        <path fill="white" d="M8.5 17.25c-2.22 0-3.13-.88-3.13-1.98 0-1.06.87-1.92 2.6-1.92 1.42 0 3.03.57 4.6 1.32-1.3 1.61-2.69 2.58-4.07 2.58z"/>
                      </svg>
                      <!-- WeChat icon -->
                      <svg v-else-if="isWechatMethod(method.value)" class="h-5 w-5 flex-shrink-0 text-[#07C160]" viewBox="0 0 24 24" fill="currentColor">
                        <path d="M9.5 4C5.36 4 2 6.69 2 10c0 1.89 1.08 3.56 2.78 4.66l-.7 2.1 2.6-1.37C7.6 15.78 8.53 16 9.5 16c.26 0 .52-.01.77-.04A5.46 5.46 0 0010 14.5c0-3.04 2.69-5.5 6-5.5.34 0 .67.03 1 .07C16.36 6.08 13.22 4 9.5 4zM6.75 7.25a1 1 0 110 2 1 1 0 010-2zm5.5 0a1 1 0 110 2 1 1 0 010-2zM16 10c-2.76 0-5 2.01-5 4.5S13.24 19 16 19c.72 0 1.4-.14 2.02-.38L20.5 20l-.5-1.88C21.23 17.12 22 15.88 22 14.5c0-2.49-2.69-4.5-6-4.5zm-2 3a.75.75 0 110 1.5.75.75 0 010-1.5zm4 0a.75.75 0 110 1.5.75.75 0 010-1.5z"/>
                      </svg>
                      <Icon v-else name="creditCard" size="md" class="flex-shrink-0 text-gray-400" />
                      <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ method.label }}</span>
                    </button>
                  </div>
                </div>

                <!-- Insufficient Balance Warning -->
                <div
                  v-if="purchaseMethod === 'balance' && insufficientBalance"
                  class="rounded-xl border border-orange-200 bg-orange-50 p-3 dark:border-orange-800/50 dark:bg-orange-900/20"
                >
                  <div class="flex items-start gap-2">
                    <Icon name="exclamationCircle" size="md" class="mt-0.5 flex-shrink-0 text-orange-600 dark:text-orange-400" />
                    <div class="text-sm text-orange-700 dark:text-orange-400">
                      <p>余额不足，当前余额 ${{ user?.balance?.toFixed(2) || '0.00' }}</p>
                      <router-link to="/wallet" class="mt-1 inline-block font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400">
                        前往充值 →
                      </router-link>
                    </div>
                  </div>
                </div>

                <!-- Submit Button -->
                <button
                  type="button"
                  :disabled="!canSubmitPurchase || purchasing"
                  class="btn btn-primary w-full py-3"
                  @click="handlePurchase"
                >
                  <svg v-if="purchasing" class="-ml-1 mr-2 h-5 w-5 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  {{ purchasing ? '处理中...' : '确认支付' }}
                </button>
              </div>
            </div>
          </div>
        </transition>
      </teleport>

      <!-- Payment Status Dialog (QR Code / Polling) -->
      <teleport to="body">
        <transition name="modal">
          <div v-if="showPaymentDialog" class="fixed inset-0 z-50 flex items-center justify-center">
            <div class="absolute inset-0 bg-black/50 backdrop-blur-sm" @click="closePaymentDialog"></div>
            <div class="relative mx-4 w-full max-w-sm rounded-2xl bg-white p-6 shadow-2xl dark:bg-dark-800">
              <div class="flex items-start justify-between">
                <h3 class="text-lg font-semibold text-gray-900 dark:text-white">完成支付</h3>
                <button type="button" class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300" @click="closePaymentDialog">
                  <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>

              <div class="mt-4 text-center">
                <!-- Payment Success State -->
                <div v-if="paymentPaidInDialog" class="space-y-4 py-4">
                  <div class="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-emerald-100 dark:bg-emerald-900/30">
                    <svg class="h-8 w-8 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
                      <path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7" />
                    </svg>
                  </div>
                  <p class="text-lg font-semibold text-emerald-600 dark:text-emerald-400">支付成功</p>
                  <p class="text-sm text-gray-500 dark:text-gray-400">订阅已激活，即将跳转...</p>
                  <button type="button" class="btn btn-primary mt-2" @click="goToSubscriptions">查看订阅</button>
                </div>

                <template v-else>
                  <!-- QR Code -->
                  <div v-if="qrCodeDataUrl" class="space-y-3">
                    <p class="text-sm text-gray-600 dark:text-gray-400">请扫描二维码完成支付</p>
                    <div class="mx-auto flex h-48 w-48 items-center justify-center rounded-xl border bg-white p-2">
                      <img :src="qrCodeDataUrl" alt="QR Code" class="h-full w-full object-contain" />
                    </div>
                  </div>

                  <!-- Payment URL -->
                  <div v-else-if="purchaseResult?.payment_url" class="space-y-3">
                    <p class="text-sm text-gray-600 dark:text-gray-400">请点击下方按钮前往支付</p>
                    <a
                      :href="purchaseResult.payment_url"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="btn btn-primary"
                    >
                      前往支付
                    </a>
                  </div>

                  <!-- Countdown & Status -->
                  <div class="mt-4 flex items-center justify-center gap-2 text-sm text-gray-500 dark:text-gray-400">
                    <svg class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    等待支付...
                    <span v-if="pollCountdown > 0">({{ pollCountdown }}s)</span>
                  </div>

                  <!-- Manual Check Button -->
                  <button
                    type="button"
                    class="mt-3 text-sm font-medium text-primary-600 hover:text-primary-700 dark:text-primary-400 dark:hover:text-primary-300"
                    :disabled="manualChecking"
                    @click="manualCheckStatus"
                  >
                    {{ manualChecking ? '查询中...' : '我已支付' }}
                  </button>
                </template>
              </div>
            </div>
          </div>
        </transition>
      </teleport>

      <!-- Order History Section -->
      <div v-if="purchaseEnabled && plans.length > 0" class="card">
        <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">订阅订单记录</h2>
        </div>
        <div class="p-6">
          <!-- Loading -->
          <div v-if="loadingOrders" class="flex items-center justify-center py-8">
            <svg class="h-6 w-6 animate-spin text-primary-500" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
          </div>

          <!-- Order List -->
          <div v-else-if="orders.length > 0" class="space-y-3">
            <div
              v-for="order in orders"
              :key="order.id"
              class="rounded-xl bg-gray-50 p-4 dark:bg-dark-800"
            >
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-4">
                  <div
                    :class="[
                      'flex h-10 w-10 items-center justify-center rounded-xl',
                      order.status === 'paid'
                        ? 'bg-emerald-100 dark:bg-emerald-900/30'
                        : order.status === 'pending'
                          ? 'bg-yellow-100 dark:bg-yellow-900/30'
                          : 'bg-gray-100 dark:bg-gray-800'
                    ]"
                  >
                    <Icon
                      name="creditCard"
                      size="md"
                      :class="
                        order.status === 'paid'
                          ? 'text-emerald-600 dark:text-emerald-400'
                          : order.status === 'pending'
                            ? 'text-yellow-600 dark:text-yellow-400'
                            : 'text-gray-400 dark:text-gray-500'
                      "
                    />
                  </div>
                  <div>
                    <p class="text-sm font-medium text-gray-900 dark:text-white">
                      {{ order.group?.name || `套餐 #${order.group_id}` }}
                    </p>
                    <p class="text-xs text-gray-500 dark:text-dark-400">
                      {{ formatDateTime(order.created_at) }}
                      · {{ getPaymentMethodLabel(order.payment_method) }}
                    </p>
                  </div>
                </div>
                <div class="text-right">
                  <p class="text-sm font-semibold text-gray-900 dark:text-white">
                    ¥{{ order.amount.toFixed(2) }}
                  </p>
                  <span
                    :class="[
                      'inline-flex rounded-full px-2.5 py-0.5 text-xs font-medium',
                      order.status === 'paid'
                        ? 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400'
                        : order.status === 'pending'
                          ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                          : order.status === 'expired'
                            ? 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'
                            : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                    ]"
                  >
                    {{ getOrderStatusLabel(order.status) }}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-else class="py-8 text-center">
            <div class="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-gray-100 dark:bg-dark-800">
              <Icon name="clock" size="xl" class="text-gray-400 dark:text-dark-500" />
            </div>
            <p class="text-sm text-gray-500 dark:text-dark-400">暂无订阅订单</p>
          </div>
        </div>
      </div>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores'
import {
  getPlans,
  purchasePlan,
  getOrders as getSubscriptionOrders,
  getOrderStatus,
  type SubscriptionPlan,
  type PurchaseResult,
  type SubscriptionOrderItem
} from '@/api/subscription-plans'
import { paymentAPI } from '@/api/payment'
import type { PaymentConfig } from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import { formatDateTime, formatDateOnly } from '@/utils/format'
import QRCode from 'qrcode'

const router = useRouter()
const authStore = useAuthStore()
const appStore = useAppStore()

const user = computed(() => authStore.user)

// Page state
const loading = ref(true)
const plans = ref<SubscriptionPlan[]>([])
const paymentConfig = ref<PaymentConfig | null>(null)

// Purchase dialog state
const showPurchaseDialog = ref(false)
const selectedPlan = ref<SubscriptionPlan | null>(null)
const purchaseMethod = ref('balance')
const purchasing = ref(false)

// Payment dialog state
const showPaymentDialog = ref(false)
const purchaseResult = ref<PurchaseResult | null>(null)
const qrCodeDataUrl = ref('')
const paymentPaidInDialog = ref(false)
const manualChecking = ref(false)
const pollCountdown = ref(300)
let currentOrderId: number | null = null
let pollTimer: ReturnType<typeof setInterval> | null = null
let countdownTimer: ReturnType<typeof setInterval> | null = null

// Order history state
const orders = ref<SubscriptionOrderItem[]>([])
const loadingOrders = ref(false)

// Computed
const purchaseEnabled = computed(() => {
  const settings = appStore.cachedPublicSettings
  // Use the new native setting key, fallback to old iframe key
  return (settings?.subscription_purchase_enabled ?? settings?.purchase_subscription_enabled) ?? false
})

const discountAmount = computed(() => {
  if (!selectedPlan.value) return 0
  if (selectedPlan.value.discounted_price != null && selectedPlan.value.discounted_price < selectedPlan.value.price) {
    return selectedPlan.value.price - selectedPlan.value.discounted_price
  }
  return 0
})

const finalPrice = computed(() => {
  if (!selectedPlan.value) return 0
  if (selectedPlan.value.discounted_price != null && selectedPlan.value.discounted_price < selectedPlan.value.price) {
    return selectedPlan.value.discounted_price
  }
  return selectedPlan.value.price
})

const insufficientBalance = computed(() => {
  return (user.value?.balance ?? 0) < finalPrice.value
})

const onlinePaymentMethods = computed(() => {
  if (!paymentConfig.value) return []
  const methods: { value: string; label: string }[] = []
  const m = paymentConfig.value.methods
  if (m.alipay) methods.push({ value: 'alipay', label: '支付宝' })
  else if (m.alipay_f2f) methods.push({ value: 'alipay_f2f', label: '支付宝当面付' })
  else if (m.epay_alipay) methods.push({ value: 'epay_alipay', label: '支付宝' })
  if (m.wechat) methods.push({ value: 'wechat', label: '微信支付' })
  else if (m.epay_wechat) methods.push({ value: 'epay_wechat', label: '微信支付' })
  return methods
})

const canSubmitPurchase = computed(() => {
  if (!purchaseMethod.value || !selectedPlan.value) return false
  if (purchaseMethod.value === 'balance' && insufficientBalance.value) return false
  return true
})

const isAlipayMethod = (method: string) => ['alipay', 'alipay_f2f', 'epay_alipay'].includes(method)
const isWechatMethod = (method: string) => ['wechat', 'epay_wechat'].includes(method)

// Plan loading
async function loadPlans() {
  try {
    plans.value = await getPlans()
  } catch (error) {
    console.error('Failed to load plans:', error)
    appStore.showError('加载套餐列表失败')
  }
}

async function loadPaymentConfig() {
  try {
    paymentConfig.value = await paymentAPI.getPaymentConfig()
  } catch {
    console.error('Failed to load payment config')
  }
}

async function loadOrders() {
  loadingOrders.value = true
  try {
    const result = await getSubscriptionOrders({ page: 1, page_size: 20 })
    orders.value = result.orders || []
  } catch {
    console.error('Failed to load subscription orders')
  } finally {
    loadingOrders.value = false
  }
}

// Purchase dialog
function openPurchaseDialog(plan: SubscriptionPlan) {
  selectedPlan.value = plan
  purchaseMethod.value = 'balance'
  showPurchaseDialog.value = true
}

function closePurchaseDialog() {
  showPurchaseDialog.value = false
  selectedPlan.value = null
}

async function handlePurchase() {
  if (!selectedPlan.value || !canSubmitPurchase.value) return
  purchasing.value = true

  try {
    const result = await purchasePlan({
      group_id: selectedPlan.value.group_id,
      payment_method: purchaseMethod.value
    })

    if (result.status === 'paid') {
      // Balance payment succeeded
      appStore.showSuccess('购买成功，订阅已激活！')
      closePurchaseDialog()
      await authStore.refreshUser()
      await loadPlans()
      await loadOrders()
      return
    }

    // Online payment - show payment dialog
    purchaseResult.value = result
    closePurchaseDialog()

    // Generate QR code
    if (result.qr_code_url) {
      try {
        qrCodeDataUrl.value = await QRCode.toDataURL(result.qr_code_url, {
          width: 256,
          margin: 2,
          color: { dark: '#000000', light: '#ffffff' }
        })
      } catch {
        qrCodeDataUrl.value = ''
      }
    } else {
      qrCodeDataUrl.value = ''
    }

    // Open form_html in new window if present
    if (result.form_html) {
      const w = window.open('', '_blank')
      if (w) {
        w.document.write(result.form_html)
        w.document.close()
      }
    }

    showPaymentDialog.value = true
    paymentPaidInDialog.value = false
    pollCountdown.value = 300
    currentOrderId = result.order_id
    startPolling(result.order_id)
  } catch (error: any) {
    appStore.showError(error.response?.data?.detail || '购买失败，请重试')
  } finally {
    purchasing.value = false
  }
}

// Polling
function startPolling(orderId: number) {
  stopPolling()
  pollTimer = setInterval(async () => {
    try {
      const result = await getOrderStatus(orderId)
      if (result.status === 'paid') {
        stopPolling()
        paymentPaidInDialog.value = true
        await authStore.refreshUser()
        await loadPlans()
        await loadOrders()
        appStore.showSuccess('支付成功，订阅已激活！')
      } else if (result.status === 'expired' || result.status === 'failed') {
        stopPolling()
        showPaymentDialog.value = false
        appStore.showError('支付已过期或失败')
      }
    } catch (e) {
      console.warn('[subscription poll] error:', e)
    }
  }, 5000)

  countdownTimer = setInterval(() => {
    pollCountdown.value--
    if (pollCountdown.value <= 0) {
      stopPolling()
      showPaymentDialog.value = false
    }
  }, 1000)
}

function stopPolling() {
  if (pollTimer) { clearInterval(pollTimer); pollTimer = null }
  if (countdownTimer) { clearInterval(countdownTimer); countdownTimer = null }
}

function closePaymentDialog() {
  stopPolling()
  showPaymentDialog.value = false
  paymentPaidInDialog.value = false
  qrCodeDataUrl.value = ''
  currentOrderId = null
}

async function manualCheckStatus() {
  if (!currentOrderId) return
  manualChecking.value = true
  try {
    const result = await getOrderStatus(currentOrderId)
    if (result.status === 'paid') {
      stopPolling()
      paymentPaidInDialog.value = true
      await authStore.refreshUser()
      await loadPlans()
      await loadOrders()
      appStore.showSuccess('支付成功，订阅已激活！')
    } else {
      appStore.showInfo('尚未检测到支付，请稍后再试')
    }
  } catch {
    appStore.showError('查询失败，请稍后再试')
  } finally {
    manualChecking.value = false
  }
}

function goToSubscriptions() {
  closePaymentDialog()
  router.push('/subscriptions')
}

// Helpers
function getPaymentMethodLabel(method: string): string {
  const map: Record<string, string> = {
    balance: '余额支付',
    alipay: '支付宝',
    alipay_f2f: '支付宝当面付',
    wechat: '微信支付',
    epay_alipay: '支付宝',
    epay_wechat: '微信支付',
    epay: '在线支付'
  }
  return map[method] || method
}

function getOrderStatusLabel(status: string): string {
  const map: Record<string, string> = {
    pending: '待支付',
    paid: '已支付',
    expired: '已过期',
    failed: '失败'
  }
  return map[status] || status
}

// Lifecycle
onMounted(async () => {
  loading.value = true
  try {
    if (!appStore.publicSettingsLoaded) {
      await appStore.fetchPublicSettings()
    }
    if (purchaseEnabled.value) {
      await Promise.all([loadPlans(), loadPaymentConfig(), loadOrders()])
    }
  } finally {
    loading.value = false
  }
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.25s ease;
}
.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}
</style>
