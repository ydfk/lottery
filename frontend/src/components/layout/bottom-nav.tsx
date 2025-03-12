import { Home, LogOut, RefreshCcw } from 'lucide-react';
import { useNavigate } from 'react-router-dom';
import { useAuthStore } from '@/store/auth-store';
import { useLotteryData } from '../../hooks/use-lottery-data';
import { cn } from '../../lib/utils';

export function BottomNav() {
  const navigate = useNavigate();
  const { logout } = useAuthStore();
  const { refresh } = useLotteryData();

  const handleRefresh = () => {
    refresh();
  };

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 border-t bg-background">
      <div className="grid grid-cols-3 h-16">
        <NavItem onClick={handleRefresh} icon={<RefreshCcw className="h-5 w-5" />} label="刷新" />
        <NavItem onClick={() => navigate('/')} icon={<Home className="h-5 w-5" />} label="首页" />
        <NavItem onClick={handleLogout} icon={<LogOut className="h-5 w-5" />} label="退出" />
      </div>
    </div>
  );
}

interface NavItemProps {
  icon: React.ReactNode;
  label: string;
  onClick: () => void;
  active?: boolean;
}

function NavItem({ icon, label, onClick, active = false }: NavItemProps) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'flex flex-col items-center justify-center gap-1',
        'transition-colors',
        active ? 'text-primary' : 'text-muted-foreground hover:text-primary'
      )}
    >
      {icon}
      <span className="text-xs">{label}</span>
    </button>
  );
}