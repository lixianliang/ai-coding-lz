# web 开发指南

## 技术栈
- 框架: Vue 3 + TypeScript
- UI 组件库: Element Plus
- 样式预处理器: SCSS
- HTTP 客户端：Flyio (基于 axios)
- 路由管理: Vue Router
- 状态管理: Pinia (推荐)
- 构建工具: Vite

## 代码结构
src/
├── apis/          # API 接口层
│   ├── types/     # 类型定义
│   └── modules/   # 接口模块
├── components/    # 公共组件
│   └── base/      # 基础组件
├── views/         # 页面组件
├── router/        # 路由配置
├── stores/        # 状态管理
├── utils/         # 工具函数
├── assets/        # 静态资源
└── styles/        # 全局样式

## 开发规范
### 组件设计原则
- 单一职责: 每个组件只负责一个特定功能
- 可复用性: 公共组件应设计为可复用
- 可维护性: 组件逻辑清晰，易于理解和修改
- 类型安全: 全面使用 TypeScript 类型定义

### 文件命名约定
- 组件文件: PascalCase，如 UserProfile.vue
- 工具文件: camelCase，如 dateUtils.ts
- 配置文件: kebab-case，如 app-config.ts
- 类型定义: PascalCase，如 User.ts

### 组件命名规范
``` vue
<!-- 推荐 -->
<UserProfile />
<BaseButton />
<DashboardHeader />

<!-- 避免 -->
<userProfile />
<base_button />
<dashboard-header />
```

## 样式设计体系
### 样式方案
```scss
// 设计令牌 (Design Tokens)
:root {
  // 颜色系统
  --color-primary: #409eff;
  --color-success: #67c23a;
  --color-warning: #e6a23c;
  --color-danger: #f56c6c;
  
  // 间距系统
  --spacing-xs: 4px;
  --spacing-sm: 8px;
  --spacing-md: 16px;
  --spacing-lg: 24px;
  
  // 字体系统
  --font-size-sm: 12px;
  --font-size-base: 14px;
  --font-size-lg: 16px;
}
```

## 样式使用优先级
- Scoped Style (主要)
- CSS 变量 (主题和设计系统)
- 内联样式 (仅动态样式)
- 全局样式 (谨慎使用)

### 样式示例
```vue
<template>
  <div class="user-card">
    <img :src="avatar" :style="avatarStyle" />
    <h3 class="user-card__name">{{ name }}</h3>
    <p class="user-card__bio">{{ bio }}</p>
  </div>
</template>

<script setup lang="ts">
// 动态样式使用内联
const avatarStyle = {
  width: '60px',
  height: '60px',
  borderRadius: '50%'
}
</script>

<style scoped lang="scss">
.user-card {
  padding: var(--spacing-md);
  border: 1px solid #ebeef5;
  border-radius: 4px;
  
  &__name {
    font-size: var(--font-size-lg);
    color: var(--color-primary);
    margin: var(--spacing-sm) 0;
  }
  
  &__bio {
    font-size: var(--font-size-base);
    color: #606266;
    line-height: 1.5;
  }
}
</style>
```

## 最佳实践
### 组件开发
```vue
<script setup lang="ts">
// 使用 Composition API
import { ref, computed } from 'vue'

interface Props {
  userId: number
  showAvatar?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showAvatar: true
})

const emit = defineEmits<{
  (e: 'update:user', user: User): void
}>()

// 响应式数据
const userData = ref<User | null>(null)

// 计算属性
const displayName = computed(() => {
  return userData.value?.name || '未知用户'
})

// 方法命名使用动词前缀
const fetchUserData = async () => {
  try {
    userData.value = await getUserById(props.userId)
  } catch (error) {
    console.error('获取用户数据失败:', error)
  }
}
</script>
```

### API 层规范
```typescript
// apis/types/user.ts
export interface User {
  id: number
  name: string
  email: string
  avatar?: string
}

// apis/modules/user.ts
import type { User } from '../types'

export const userApi = {
  // 获取用户详情
  async getById(id: number): Promise<User> {
    const response = await fetch(`/api/users/${id}`)
    return response.json()
  },
  
  // 更新用户信息
  async update(user: Partial<User>): Promise<User> {
    const response = await fetch(`/api/users/${user.id}`, {
      method: 'PATCH',
      body: JSON.stringify(user)
    })
    return response.json()
  }
}
```

### 路由配置
```typescript
// router/index.ts
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'Home',
    component: () => import('@/views/Home.vue'),
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/user/:id',
    name: 'UserProfile',
    component: () => import('@/views/UserProfile.vue'),
    props: true
  }
]
```

## 代码质量

### 提交规范
- feat: 新功能
- fix: 修复问题
- docs: 文档更新
- style: 代码格式调整
- refactor: 代码重构
- test: 测试相关

## 代码检查
- 使用 ESLint + Prettier 统一代码风格
- 提交前自动进行代码检查
