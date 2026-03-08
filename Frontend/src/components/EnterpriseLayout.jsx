import { Outlet, NavLink, useNavigate } from 'react-router';
import { useEnterpriseAuth } from '../lib/EnterpriseAuthContext';
import toast from 'react-hot-toast';

const navItems = [
    {
        to: '/enterprise/dashboard',
        label: 'Overview',
        icon: (
            <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6" />
            </svg>
        ),
    },
    {
        to: '/enterprise/keys',
        label: 'API Keys',
        icon: (
            <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
            </svg>
        ),
    },
    {
        to: '/enterprise/usage',
        label: 'Usage & Billing',
        icon: (
            <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
        ),
    },
    {
        to: '/enterprise/docs',
        label: 'API Docs',
        icon: (
            <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
            </svg>
        ),
    },
    {
        to: '/enterprise/profile',
        label: 'Account',
        icon: (
            <svg xmlns="http://www.w3.org/2000/svg" className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
            </svg>
        ),
    },
];

export default function EnterpriseLayout() {
    const { entUser, logout } = useEnterpriseAuth();
    const navigate = useNavigate();

    const handleLogout = () => {
        logout();
        toast.success('Signed out of enterprise portal');
        navigate('/enterprise');
    };

    return (
        <div className="min-h-screen flex bg-slate-950 font-mono" style={{ fontFamily: "'DM Mono', 'Fira Code', monospace" }}>
            {/* Sidebar */}
            <aside className="w-64 flex flex-col bg-slate-900 border-r border-slate-800 fixed h-full z-10">
                {/* Logo */}
                <div className="px-6 py-6 border-b border-slate-800">
                    <div className="flex items-center gap-3">
                        <div className="w-8 h-8 rounded-lg bg-violet-500 flex items-center justify-center">
                            <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
                            </svg>
                        </div>
                        <div>
                            <p className="text-xs font-bold tracking-widest text-violet-400 uppercase">MAP Enterprise</p>
                            <p className="text-xs text-slate-600 tracking-wide">API Platform</p>
                        </div>
                    </div>
                </div>

                {/* Nav */}
                <nav className="flex-1 px-3 py-4 space-y-1">
                    {navItems.map((item) => (
                        <NavLink
                            key={item.to}
                            to={item.to}
                            className={({ isActive }) =>
                                `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all duration-150 ${isActive
                                    ? 'bg-violet-500/10 text-violet-400 border border-violet-500/20'
                                    : 'text-slate-500 hover:text-slate-200 hover:bg-slate-800'
                                }`
                            }
                        >
                            {item.icon}
                            <span>{item.label}</span>
                        </NavLink>
                    ))}
                </nav>

                {/* Enterprise badge */}
                <div className="px-4 py-3 mx-3 mb-3 rounded-lg bg-violet-500/5 border border-violet-500/15">
                    <p className="text-xs text-violet-400 font-bold">Enterprise Plan</p>
                    <p className="text-xs text-slate-600 mt-0.5">{entUser?.company_name || entUser?.name || 'Your Company'}</p>
                </div>

                {/* User footer */}
                <div className="px-4 py-4 border-t border-slate-800">
                    <div className="flex items-center gap-3 mb-3">
                        <div className="w-8 h-8 rounded-full bg-violet-500/20 border border-violet-500/30 flex items-center justify-center text-violet-400 font-bold text-sm">
                            {(entUser?.name || entUser?.company_name || 'E')[0]?.toUpperCase()}
                        </div>
                        <div className="flex-1 min-w-0">
                            <p className="text-xs font-semibold truncate text-slate-200">{entUser?.name || entUser?.company_name}</p>
                            <p className="text-xs text-slate-600 truncate">{entUser?.email}</p>
                        </div>
                    </div>
                    <button
                        onClick={handleLogout}
                        className="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-xs text-rose-400 hover:bg-rose-500/10 transition-colors"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
                        </svg>
                        Sign out
                    </button>
                </div>
            </aside>

            {/* Main */}
            <main className="flex-1 ml-64 min-h-screen">
                <Outlet />
            </main>
        </div>
    );
}