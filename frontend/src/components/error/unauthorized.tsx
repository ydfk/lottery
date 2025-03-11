import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../ui/button";

export default function Unauthorized() {
  const navigate = useNavigate();

  return (
    <div className="min-h-screen bg-white dark:bg-gray-900 text-gray-900 dark:text-white flex flex-col items-center justify-center p-4">
      <div className="w-full max-w-md space-y-8 text-center">
        <h1 className="text-4xl font-bold">401 Unauthorized</h1>
        <p className="text-lg text-gray-500 dark:text-gray-400">
          您未登录或登录已过期，请重新登录
        </p>
        <Button onClick={() => navigate("/auth/login")}>
          返回登录页面
        </Button>
      </div>
    </div>
  );
}