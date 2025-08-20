// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../store/auth';

import Login from '../views/Login.vue';
import Dashboard from '../views/Dashboard.vue';
import Connections from '../views/Connections.vue';
import Users from '../views/Users.vue';
import TransactionList from '../views/TransactionList.vue';
import InvoiceDetail from '../views/InvoiceDetail.vue';
import CustomerDetail from '../views/CustomerDetail.vue';
import ProductDetail from '../views/ProductDetail.vue';
import SellerDetail from '../views/SellerDetail.vue';
import Settings from '../views/Settings.vue';

const routes = [
    {
        path: '/settings',
        name: 'Settings',
        component: Settings,
        meta: {requiresAuth: true}
    },
    {
        path: '/login',
        name: 'Login',
        component: Login,
        meta: { requiresAuth: false }
    },
    {
        path: '/dashboard',
        name: 'Dashboard',
        component: Dashboard,
        meta: { requiresAuth: true }
    },
    {
        path: '/connections',
        name: 'Connections',
        component: Connections,
        meta: { requiresAuth: true }
    },
    {
        path: '/users',
        name: 'Users',
        component: Users,
        meta: { requiresAuth: true }
    },
    {
        // Esta es una ruta dinámica. ':type' será un parámetro en la URL
        // que nos dirá qué tipo de transacciones mostrar (ej: /transactions/invoices).
        path: '/transactions/:type',
        name: 'TransactionList', // <-- Este es el nombre que el router estaba buscando.
        component: TransactionList,
        meta: { requiresAuth: true },
        // 'props: true' permite que los parámetros de la ruta (como 'type')
        // se pasen como propiedades (props) al componente, lo cual es una buena práctica.
        props: true
    },
    {
        // Ruta para el detalle de una factura. ':id' será el número de la factura.
        // ej: /invoice/123
        path: '/invoice/:id',
        name: 'InvoiceDetail',
        component: InvoiceDetail,
        meta: { requiresAuth: true },
        props: true
    },
    {
        path: '/customer/:id',
        name: 'CustomerDetail',
        component: CustomerDetail,
        meta: {requiresAuth: true},
        props: true
    },
    {
        path: '/seller/:id',
        name: 'SellerDetail',
        component: SellerDetail,
        meta: {requiresAuth: true},
        props: true
    },
    {
        path: '/product/:id',
        name: 'ProductDetail',
        component: ProductDetail,
        meta: {requiresAuth: true},
        props: true
    },
    {
        path: '/',
        redirect: () => {
            const authStore = useAuthStore();
            return authStore.isAuthenticated ? '/dashboard' : '/login';
        }
    }
];

const router = createRouter({
    history: createWebHistory(),
    routes,
});

// El 'Navigation Guard' (guardia de navegación) no necesita cambios.
router.beforeEach((to, from, next) => {
    const authStore = useAuthStore();
    const requiresAuth = to.matched.some(record => record.meta.requiresAuth);

    if (requiresAuth && !authStore.isAuthenticated) {
        next('/login');
    } else if (to.name === 'Login' && authStore.isAuthenticated) {
        next('/dashboard');
    } else {
        next();
    }
});

export default router;