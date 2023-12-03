'use client';

import {memo, useCallback, useState, MouseEvent} from "react";
import {CopyIcon, EditIcon, OptionsIcon, TrashIcon} from "@/app/components/Icons";
import {tw} from "@/app/utils/style";

interface ProductProps {
    name: string;
    location: string;
    image: string;
    removeProduct: () => void;
}

export const Product = memo(function Product({name, location, image, removeProduct}: ProductProps) {
    const [optionsOpen, setOptionsOpen] = useState(false);
    const [{x, y}, setCoords] = useState({x: '0px', y: '0px'});

    const toggleOptions = useCallback((ev: MouseEvent<HTMLButtonElement>) => {
        setOptionsOpen(prevState => !prevState);
        setCoords(prev => {
            if (prev.x !== '0px' && prev.y !== '0px') {
                return prev;
            }

            const {x, width, height, y} = ev.currentTarget.getBoundingClientRect();

            const middleX = x + (width / 2);
            const middleY = y + (height / 2);

            return {
                x: `${window.innerWidth - middleX}px`,
                y: `${window.innerHeight - middleY}px`
            }
        })
    }, []);

    return (
        <div className="flex basis-[0%] md:flex-col group max-md:items-center relative">
            <img
                loading="lazy"
                src={image}
                alt={name}
                className="aspect-[1.57] rounded-sm shadow-black object-cover object-center w-1/5 md:w-full shadow-sm overflow-hidden group-hover:border-2 border-zinc-700 dark:border-neutral-50 transition-all duration-200 ease-in-out group-hover:shadow-lg"
            />
            <div className="justify-between flex gap-0 md:mt-3.5 max-md:ml-4 max-md:grow max-md:items-center">
                <div className="flex grow basis-[0%] flex-col">
                    <div className="text-black dark:text-white text-sm font-semibold">{name}</div>
                    <div className="text-zinc-500 dark:text-zinc-300 text-sm mt-1.5">{location}</div>
                </div>
                <button className={'h-fit'} onClick={toggleOptions}>
                    <OptionsIcon className="text-zinc-500 dark:text-zinc-300" />
                </button>
            </div>
            <div
                className={
                    tw('fixed px-2 bg-white dark:bg-zinc-700 shadow-sm rounded-md flex items-start justify-center flex-col border border-[color:var(--border,#EBEBEB)] dark:border-[color:var(--border,#383944)]', {
                        hidden: !optionsOpen
                     })
                }
                style={
                    {
                        right: x,
                        bottom: y
                    }
                }
            >
                <button className="p-2 flex items-center text-zinc-500 dark:text-zinc-300 text-sm font-semibold">
                    <EditIcon className="mr-2" />
                    <span>Éditer</span>
                </button>
                <button className="p-2 flex items-center text-zinc-500 dark:text-zinc-300 text-sm font-semibold">
                    <CopyIcon className="mr-2" />
                    <span>Dupliquer</span>
                </button>
                <button className="p-2 flex items-center text-red-500 dark:text-red-300 text-sm font-semibold" onClick={removeProduct}>
                    <TrashIcon className="mr-2" />
                    <span>Supprimer</span>
                </button>
            </div>
        </div>
    )
})
