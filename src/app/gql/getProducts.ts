import {gql, useMutation, useQuery} from '@apollo/client';
import {useCallback, useMemo, useState} from "react";

const GET_MY_PROJECTS = gql`
    query {
        projects(filter: MINE){
            _id
            properties {
                name
                town
            }
            image {
                url {
                    original
                }
            }
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

export const useMyProjects = () => {
    const [projects, setProjects] = useState<Project[]>([]);
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
