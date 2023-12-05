import {gql, useMutation, useQuery} from '@apollo/client';
import {useCallback, useMemo, useState} from "react";

const GET_MY_PROJECTS = gql`
    query projects($filter: MINE){
    _id
    directory {
    _id
    }
    properties {
    name
    description
    client
    startDate
    endDate
    address
    zipcode
    town
    state
    country
    phone
    mobile
    email
    }
    }
`;

export const DELETE_PROJECT = gql`
    mutation($projects: [ID!]!) {
        deleteProjects(projects: $projects)
    }
`;

interface Project {
    _id: string;
    properties: {
        name: string;
        town: string;
    };
    image: {
        url: {
            original: string;
        }
    }
}

const defaultProducts: Project[] = [
    {
        _id: "1",
        properties: {
            name: "Lumos",
            town: "Brest"
        },
        image: {
            url: {
                original: 'https://s3-alpha-sig.figma.com/img/9de1/2203/7caf1234f6c034a5527db21e68e1fa87?Expires=1702252800&Signature=MZrgnz~lOA6gGZUeZNEd73RDl5NQqYjrvffZABkxS4CiA0EYiHYbWF-SgnYLq3-C4B3H1w2i32uQZIKbdHBo2aExK~Zr3m2-x2XuiZ05CdV5xcUGMFVR66SzFW1gogGwBnvYUZqF2gCvrmwIMfLKkNTaDyFp4UcFy6zH0hn7JZcktiePunKbiGbXtt5pLmrnxhQRUN17UjahswzvnkAGQMvSiLXf2dbjEMN8mhP9mstSxclAANcvgElT867Txgl9D~Vvz5rYI2ZDmf8waWne-EAMw5mcWvRxsExXHGX1znz-MC~NOUIw9qz689yPTZ4JtTjaiwOHkuqkDnCajWBB0w__&Key-Pair-Id=APKAQ4GOSFWCVNEHN3O4',
            }
        }
    },
    {
        _id: "2",
        properties: {
            name: "Accio",
            town: "Vannes"
        },
        image: {
            url: {
                original: 'https://s3-alpha-sig.figma.com/img/0df9/2f39/820e40638868ae2d8b40796e00a0f271?Expires=1702252800&Signature=jiK5wHdNHeNCKWXoLt9HHL9B5P3ja5iBmBYzhWA9YlwLWIvQHPAojppCLhlFV~M-fwEzcSmOfZIZpMUuTErSHlRmxXdWAUBT2CxIrlxawd22mz6oiJ03IkI78oRW8LHLbK2MZDMFcQSjcqoh-XDH-81diJ~FmIQ04YYIWmxjcxCmNDkhogz0GeMJfic4tTMqn6dys4y2fpOEMcUVymmiyra7QxQsO-pkJl8ZirZI8PE9jTU~WEgVdetBEu2mN7Wc1y8SHKBPD54WBG-Gx8Z6mMQp3rBReYzaTpwGNfkd8AR8GMVE94jE7ca23k6PFM~IJ3q6KePCSudH81VUjK8kgg__&Key-Pair-Id=APKAQ4GOSFWCVNEHN3O4',
            }
        }
    },
    {
        _id: "3",
        properties: {
            name: "Diffinito",
            town: "Rennes"
        },
        image: {
            url: {
                original: 'https://s3-alpha-sig.figma.com/img/9de1/2203/7caf1234f6c034a5527db21e68e1fa87?Expires=1702252800&Signature=MZrgnz~lOA6gGZUeZNEd73RDl5NQqYjrvffZABkxS4CiA0EYiHYbWF-SgnYLq3-C4B3H1w2i32uQZIKbdHBo2aExK~Zr3m2-x2XuiZ05CdV5xcUGMFVR66SzFW1gogGwBnvYUZqF2gCvrmwIMfLKkNTaDyFp4UcFy6zH0hn7JZcktiePunKbiGbXtt5pLmrnxhQRUN17UjahswzvnkAGQMvSiLXf2dbjEMN8mhP9mstSxclAANcvgElT867Txgl9D~Vvz5rYI2ZDmf8waWne-EAMw5mcWvRxsExXHGX1znz-MC~NOUIw9qz689yPTZ4JtTjaiwOHkuqkDnCajWBB0w__&Key-Pair-Id=APKAQ4GOSFWCVNEHN3O4',
            }
        }
    },
    {
        _id: "4",
        properties: {
            name: "Incendio",
            town: "Saint-Brieuc"
        },
        image: {
            url: {
                original: 'https://s3-alpha-sig.figma.com/img/0df9/2f39/820e40638868ae2d8b40796e00a0f271?Expires=1702252800&Signature=jiK5wHdNHeNCKWXoLt9HHL9B5P3ja5iBmBYzhWA9YlwLWIvQHPAojppCLhlFV~M-fwEzcSmOfZIZpMUuTErSHlRmxXdWAUBT2CxIrlxawd22mz6oiJ03IkI78oRW8LHLbK2MZDMFcQSjcqoh-XDH-81diJ~FmIQ04YYIWmxjcxCmNDkhogz0GeMJfic4tTMqn6dys4y2fpOEMcUVymmiyra7QxQsO-pkJl8ZirZI8PE9jTU~WEgVdetBEu2mN7Wc1y8SHKBPD54WBG-Gx8Z6mMQp3rBReYzaTpwGNfkd8AR8GMVE94jE7ca23k6PFM~IJ3q6KePCSudH81VUjK8kgg__&Key-Pair-Id=APKAQ4GOSFWCVNEHN3O4',
            }
        }
    },
    {
        _id: "5",
        properties: {
            name: "Reducto",
            town: "Saint-Georges-De-Reintembault"
        },
        image: {
            url: {
                original: 'https://s3-alpha-sig.figma.com/img/9de1/2203/7caf1234f6c034a5527db21e68e1fa87?Expires=1702252800&Signature=MZrgnz~lOA6gGZUeZNEd73RDl5NQqYjrvffZABkxS4CiA0EYiHYbWF-SgnYLq3-C4B3H1w2i32uQZIKbdHBo2aExK~Zr3m2-x2XuiZ05CdV5xcUGMFVR66SzFW1gogGwBnvYUZqF2gCvrmwIMfLKkNTaDyFp4UcFy6zH0hn7JZcktiePunKbiGbXtt5pLmrnxhQRUN17UjahswzvnkAGQMvSiLXf2dbjEMN8mhP9mstSxclAANcvgElT867Txgl9D~Vvz5rYI2ZDmf8waWne-EAMw5mcWvRxsExXHGX1znz-MC~NOUIw9qz689yPTZ4JtTjaiwOHkuqkDnCajWBB0w__&Key-Pair-Id=APKAQ4GOSFWCVNEHN3O4',
            }
        }
    },
];

export const useMyProjects = () => {
    const [projects, setProjects] = useState<Project[]>([...defaultProducts, ...defaultProducts, ...defaultProducts, ...defaultProducts, ...defaultProducts]);
    const {data, loading, error} = useQuery(GET_MY_PROJECTS);
    const [deleteProject] = useMutation(DELETE_PROJECT);

    useMemo(() => {
        if (data && data.projects) {
            setProjects(data.projects);
        }
    }, [data]);

    const removeProject = useCallback((index: number, project_id: string) => {
        const newProjects = [...projects];
        newProjects.splice(index, 1);
        setProjects(newProjects);
        void deleteProject({
            variables: {
                projects: [project_id],
            },
        });
    }, [deleteProject, projects]);

    return {
        removeProject,
        projects,
        loading,
        error,
    };
}
