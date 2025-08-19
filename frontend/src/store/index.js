// src/router/index.js
import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../store/auth';

// 1. Importamos las vistas existentes y las NUEVAS vistas que usará el drilldown.
import Login from '../views/Login.vue';
import Dashboard from '../views/Dashboard.vue';
import Connections from '../views/Connections.vue';
import Users from '../views/Users.vue';
import TransactionList from '../views/TransactionList.vue'; // <-- Asegúrate de que esta línea esté presente
import InvoiceDetail from '../views/InvoiceDetail.vue';   // <-- Y esta también

const routes = [
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

    // --- RUTAS PARA DRILLDOWN (Añadidas) ---
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
    // --- FIN DE NUEVAS RUTAS ---

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