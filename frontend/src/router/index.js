// frontend/src/router/index.js
import { createRouter, createWebHistory } from 'vue-router';
import { useAuthStore } from '../store/auth'; // Importaremos el store de autenticación

// Importa tus vistas (las crearemos en los siguientes pasos)
import Login from '../views/Login.vue';
import Dashboard from '../views/Dashboard.vue';
import Connections from '../views/Connections.vue';
import Users from '../views/Users.vue';

const routes = [
    {
        path: '/login',
        name: 'Login',
        component: Login,
        meta: { requiresAuth: false } // Esta ruta no requiere autenticación
    },
    {
        path: '/dashboard',
        name: 'Dashboard',
        component: Dashboard,
        meta: { requiresAuth: true } // Esta ruta está protegida
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
    // Redirección por defecto: si el usuario está logueado, va al dashboard, si no, al login.
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

// Guardia de Navegación Global (Navigation Guard)
// Esto se ejecuta ANTES de cada cambio de ruta. Es nuestro punto de control de seguridad.
router.beforeEach((to, from, next) => {
    const authStore = useAuthStore();
    const requiresAuth = to.matched.some(record => record.meta.requiresAuth);

    if (requiresAuth && !authStore.isAuthenticated) {
        // Si la ruta requiere autenticación y el usuario no está logueado,
        // se le redirige a la página de login.
        next('/login');
    } else if (to.name === 'Login' && authStore.isAuthenticated) {
        // Si el usuario ya está logueado e intenta ir a /login,
        // se le redirige al dashboard.
        next('/dashboard');
    } else {
        // En cualquier otro caso, se permite la navegación.
        next();
    }
});

export default router;