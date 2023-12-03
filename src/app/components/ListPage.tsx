'use client';

import {Product} from "@/app/components/Product";
import {ThemeButton} from "@/app/components/ThemeButton";
import {DisconnectButton} from "@/app/components/DisconnectButton";
import {PlusIcon} from "@/app/components/Icons";
import {useAuthContext} from "@/app/components/AuthContext";
import {useMyProjects} from "@/app/gql/getProducts";

export const ListPage = () => {
    const {auth} = useAuthContext();
    const { projects, removeProject } = useMyProjects();

    return (
        <div className="items-start flex flex-col text-black dark:text-white h-screen overflow-y-clip">
            <div className="justify-between items-center self-stretch border border-[color:var(--border,#EBEBEB)] dark:border-[color:var(--border,#383944)] flex w-full gap-5 px-8 py-5 border-solid max-md:max-w-full max-md:flex-wrap max-md:px-5">
                <div className="text-base font-semibold grow whitespace-nowrap my-auto">
                    Bienvenu {auth?.properties.firstname} {auth?.properties.lastname.toUpperCase()}
                </div>
                <div className="items-stretch self-stretch flex justify-between gap-3">
                    <ThemeButton/>
                    <DisconnectButton/>
                </div>
            </div>
            <div className="justify-between items-center self-stretch flex w-full gap-5 px-8 py-5 max-md:max-w-full max-md:flex-wrap max-md:px-5">
                <h2 className="text-xl font-semibold grow whitespace-nowrap my-auto">
                    Projets
                </h2>
                <div className="flex text-white justify-between items-center shadow-sm bg-indigo-500 self-stretch gap-1.5 px-4 py-2 rounded-full">
                    <PlusIcon className="w-4 h-4" strokeWidth={2}/>
                    <button className="text-white text-sm font-semibold self-center grow whitespace-nowrap my-auto">
                        Nouveau projet
                    </button>
                </div>
            </div>
            <div className="w-full grid grid-cols-1 md:grid-cols-4 lg:grid-cols-5 gap-5 px-8 max-md:px-5 h-[80vh] md:h-[85vh] overflow-y-scroll pb-12">
                {
                    projects.map((project, index) => (
                        <Product
                            key={`${project._id}-${index}`}
                            name={project.properties.name}
                            location={project.properties.town}
                            image={project.image.url.original}
                            removeProduct={removeProject.bind(null, index, project._id)}
                        />
                    ))
                }
            </div>
        </div>
    );
}
