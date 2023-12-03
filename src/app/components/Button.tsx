import {memo, ReactNode} from "react";
import {tw} from "@/app/utils/style";

interface RoundButtonProps {
    className?: string;
    Icon: ReactNode;
    handleClick: () => void;
    tooltip: string;
}

export const RoundButton = memo(function RoundButton({className, Icon, handleClick, tooltip}: RoundButtonProps) {
    return (
        <button
            title={tooltip}
            className={tw('justify-center items-center shadow-sm bg-white dark:bg-zinc-700 dark:text-white text-zinc-700 flex aspect-square flex-col w-9 h-9 p-2 rounded-full', className)}
            onClick={handleClick}
        >
            {Icon}
        </button>
    )
});
